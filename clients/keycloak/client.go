package keycloak

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
)

type StatusError struct {
	Method     string
	Path       string
	StatusCode int
	Body       string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("Keycloak answered %d to %s %s: %s", e.StatusCode, e.Method, e.Path, e.Body)
}

func isNotFound(err error) bool {
	var statusErr *StatusError

	return errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusNotFound
}

type Client struct {
	Config  *Config
	Session *session.Session

	httpClient  *http.Client
	accessToken string
	tokenExpiry time.Time
}

// NewClient constructs a Keycloak admin client from an already loaded configuration
func NewClient(keycloakConfig *Config, session *session.Session) (*Client, error) {
	if keycloakConfig == nil {
		return nil, fmt.Errorf("a Keycloak client needs a config")
	}

	keycloakConfig.applyDefaults()

	if keycloakConfig.Host == "" {
		return nil, fmt.Errorf("host is empty, the Keycloak server address is required")
	}

	if keycloakConfig.Realm == "" {
		return nil, fmt.Errorf("realm is empty, the realm to administer is required")
	}

	if keycloakConfig.AdminUser == "" || keycloakConfig.AdminPassword == "" {
		return nil, fmt.Errorf("adminUser and adminPassword must both be set, the client administers "+
			"the %s realm on the caller's behalf", keycloakConfig.Realm)
	}

	normalizedHost := strings.TrimRight(keycloakConfig.Host, "/")
	if !strings.HasPrefix(normalizedHost, "https://") && !strings.HasPrefix(normalizedHost, "http://") {
		normalizedHost = "https://" + normalizedHost
	}

	parsedHost, err := url.Parse(normalizedHost)
	if err != nil {
		return nil, fmt.Errorf("parsing Keycloak host %s: %w", keycloakConfig.Host, err)
	}
	if parsedHost.Host == "" {
		return nil, fmt.Errorf("Keycloak host %s has no host", keycloakConfig.Host)
	}

	keycloakConfig.Host = normalizedHost

	return &Client{
		Config:  keycloakConfig,
		Session: session,
		httpClient: &http.Client{
			Timeout: defaults.OneMinuteTimeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: keycloakConfig.Insecure},
			},
		},
	}, nil
}

// NewClientFromConfigKey constructs a Keycloak admin client after it reads the given configuration key
func NewClientFromConfigKey(configKey string, session *session.Session) (*Client, error) {
	keycloakConfig := new(Config)
	config.LoadConfig(configKey, keycloakConfig)

	return NewClient(keycloakConfig, session)
}

// Realm returns the realm this client administers
func (c *Client) Realm() string {
	return c.Config.Realm
}

func (c *Client) token() (string, error) {
	if c.accessToken != "" && time.Now().Add(defaults.TenSecondTimeout).Before(c.tokenExpiry) {
		return c.accessToken, nil
	}

	form := url.Values{
		"grant_type": {"password"},
		"client_id":  {c.Config.AdminClientID},
		"username":   {c.Config.AdminUser},
		"password":   {c.Config.AdminPassword},
	}

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", c.Config.Host, c.Config.AdminRealm)

	resp, err := c.httpClient.PostForm(tokenURL, form)
	if err != nil {
		return "", fmt.Errorf("requesting a Keycloak admin token: %w", err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading the Keycloak admin token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Keycloak refused an admin token for %s in realm %s, it answered %d: %s",
			c.Config.AdminUser, c.Config.AdminRealm, resp.StatusCode, payload)
	}

	var grant struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(payload, &grant); err != nil {
		return "", fmt.Errorf("decoding the Keycloak admin token response: %w", err)
	}

	c.accessToken = grant.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(grant.ExpiresIn) * time.Second)

	return c.accessToken, nil
}

func (c *Client) adminPath(format string, args ...any) string {
	return fmt.Sprintf("/admin/realms/%s", c.Config.Realm) + fmt.Sprintf(format, args...)
}

func (c *Client) do(method, path string, body, out any) error {
	token, err := c.token()
	if err != nil {
		return err
	}

	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding the body of %s %s: %w", method, path, err)
		}

		payload = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, c.Config.Host+path, payload)
	if err != nil {
		return fmt.Errorf("building the request for %s %s: %w", method, path, err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending %s %s to Keycloak: %w", method, path, err)
	}
	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading the response to %s %s: %w", method, path, err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return &StatusError{Method: method, Path: path, StatusCode: resp.StatusCode, Body: string(response)}
	}

	if out == nil || len(response) == 0 {
		return nil
	}

	if err := json.Unmarshal(response, out); err != nil {
		return fmt.Errorf("decoding the response to %s %s: %w", method, path, err)
	}

	return nil
}

func (c *Client) getPublic(path string) ([]byte, error) {
	resp, err := c.httpClient.Get(c.Config.Host + path)
	if err != nil {
		return nil, fmt.Errorf("sending GET %s to Keycloak: %w", path, err)
	}
	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading the response to GET %s: %w", path, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Keycloak answered %d to GET %s: %s", resp.StatusCode, path, response)
	}

	return response, nil
}
