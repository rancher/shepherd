package oidc

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	oidcext "github.com/rancher/shepherd/extensions/auth/oidc"
	sheptoken "github.com/rancher/shepherd/extensions/token"
	"github.com/rancher/shepherd/pkg/clientbase"
	"github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/sirupsen/logrus"
)

type APIClient struct {
	rancherURL       string
	httpClient       *http.Client
	noRedirectClient *http.Client
	Config           *Config
	session          *session.Session
}

// NewAPIClient constructs an APIClient and loads the "oidc" config block.
func NewAPIClient(rancherURL string, session *session.Session) *APIClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint:gosec
	}
	normalizedURL := strings.TrimRight(rancherURL, "/")
	if !strings.HasPrefix(normalizedURL, "https://") && !strings.HasPrefix(normalizedURL, "http://") {
		normalizedURL = "https://" + normalizedURL
	}
	oidcConfig := new(Config)
	config.LoadConfig(ConfigurationFileKey, oidcConfig)
	return &APIClient{
		rancherURL: normalizedURL,
		Config:     oidcConfig,
		session:    session,
		httpClient: &http.Client{Transport: transport},
		noRedirectClient: &http.Client{
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// GetDiscovery fetches the OIDC provider metadata document.
func (c *APIClient) GetDiscovery() (*clientbase.Response, map[string]interface{}, error) {
	return oidcext.Discovery(c.httpClient, c.rancherURL)
}

// RefreshAccessToken exchanges a refresh_token for a new TokenSet.
func (c *APIClient) RefreshAccessToken(refreshToken, clientID, clientSecret string) (*oidcext.TokenSet, error) {
	return oidcext.RefreshAccessToken(c.httpClient, c.rancherURL, refreshToken, clientID, clientSecret)
}

// RawRequest executes an HTTP request against a Rancher API path.
func (c *APIClient) RawRequest(method, path, authHeader string) (*clientbase.Response, error) {
	headers := map[string]string{}
	if authHeader != "" {
		headers["Authorization"] = authHeader
	}
	return clientbase.Do(c.httpClient, method, c.rancherURL+path, nil, headers)
}

// CompleteAuthCodeFlow drives the headless PKCE authorization-code flow and returns a TokenSet.
func (c *APIClient) CompleteAuthCodeFlow(clientID, clientSecret, redirectURI, scopes, username, password string) (*oidcext.TokenSet, error) {
	const maxAttempts = 5
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		tokenResp, err := sheptoken.GenerateUserToken(&management.User{Username: username, Password: password}, strings.TrimPrefix(c.rancherURL, "https://"))
		if err != nil {
			return nil, fmt.Errorf("step 1 (Rancher login): %w", err)
		}
		rancherToken := tokenResp.Token
		pkce, err := oidcext.GeneratePKCE()
		if err != nil {
			return nil, fmt.Errorf("step 2 (PKCE generation): %w", err)
		}
		state, err := randomState(12)
		if err != nil {
			return nil, fmt.Errorf("step 2 (state generation): %w", err)
		}
		nonce, err := randomState(12)
		if err != nil {
			return nil, fmt.Errorf("step 2 (nonce generation): %w", err)
		}
		params := url.Values{
			"response_type":         {"code"},
			"client_id":             {clientID},
			"redirect_uri":          {redirectURI},
			"scope":                 {scopes},
			"state":                 {state},
			"nonce":                 {nonce},
			"code_challenge":        {pkce.Challenge},
			"code_challenge_method": {"S256"},
		}
		authURL := c.rancherURL + oidcext.OIDCAuthPath + "?" + params.Encode()
		authResp, err := clientbase.Do(c.noRedirectClient, "GET", authURL, nil, map[string]string{
			"Authorization": "Bearer " + rancherToken,
		})
		if err != nil {
			return nil, fmt.Errorf("step 3 (auth endpoint GET): %w", err)
		}
		if authResp.StatusCode != http.StatusFound {
			return nil, fmt.Errorf("step 3 expected 302 from auth endpoint, got %d: %s",
				authResp.StatusCode, authResp.Body)
		}
		location := authResp.Header.Get("Location")
		if location == "" {
			return nil, fmt.Errorf("step 4: auth endpoint returned 302 but no Location header")
		}
		if strings.Contains(location, "/dashboard/auth/login") {
			logrus.Infof("step 4: auth endpoint redirected to dashboard login — Rancher session token was rejected, retrying")
			continue
		}
		redirectParsed, err := url.Parse(location)
		if err != nil {
			return nil, fmt.Errorf("step 4: parsing Location URL %q: %w", location, err)
		}
		authCode := redirectParsed.Query().Get("code")
		if authCode == "" {
			return nil, fmt.Errorf("step 4: no 'code' parameter in redirect Location %q", location)
		}
		if redirectParsed.Query().Get("state") != state {
			return nil, fmt.Errorf("step 4: state mismatch")
		}
		body := url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {authCode},
			"code_verifier": {pkce.Verifier},
			"client_id":     {clientID},
			"client_secret": {clientSecret},
			"redirect_uri":  {redirectURI},
		}
		resp, err := clientbase.Do(c.httpClient, "POST", c.rancherURL+oidcext.OIDCTokenPath, body, map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		})
		if err != nil {
			return nil, fmt.Errorf("token exchange POST: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("token exchange returned %d: %s", resp.StatusCode, resp.Body)
		}
		var ts oidcext.TokenSet
		if err := json.Unmarshal(resp.Body, &ts); err != nil {
			return nil, fmt.Errorf("parsing token response: %w", err)
		}
		if ts.AccessToken == "" {
			return nil, fmt.Errorf("token response missing access_token field")
		}
		return &ts, nil
	}
	return nil, fmt.Errorf("PKCE auth flow failed after %d attempts", maxAttempts)
}

func randomState(n int) (string, error) {
	raw := make([]byte, n)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generating random state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
