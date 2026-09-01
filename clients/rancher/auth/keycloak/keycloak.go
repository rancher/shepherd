package keycloak

import (
	"fmt"
	//"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"

	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/sirupsen/logrus"
)

// KeycloakOperations defines the interface for interacting with the Keycloak Auth Provider
type KeycloakOperations interface {
	Enable() error
	Disable() error
	Update(existing *management.AuthConfig, updates interface{}) (*management.AuthConfig, error)
}

const (
	resourceID           = "keycloak"
	resourceType         = "keyCloakConfig"
	ConfigurationFileKey = "keycloak"
)

type KeycloakClient struct {
	client  *management.Client
	session *session.Session
	Config  *Config
}

func NewKeycloak(client *management.Client, session *session.Session) (*KeycloakClient, error) {
	keycloakConfig := new(Config)
	config.LoadConfig(ConfigurationFileKey, keycloakConfig)

	return &KeycloakClient{
		client:  client,
		session: session,
		Config:  keycloakConfig,
	}, nil
}

func (k *KeycloakClient) Enable() error {
	existingConfig, err := k.client.AuthConfig.ByID(resourceID)
	if err != nil {
		return err
	}

	updateInput := k.newEnableInputFromConfig()

	var result management.AuthConfig
	err = k.client.Ops.DoUpdate(resourceType, &existingConfig.Resource, updateInput, &result)
	if err != nil {
		return err
	}

	k.session.RegisterCleanupFunc(func() error {
		return k.Disable()
	})

	return nil
}

func (k *KeycloakClient) Disable() error {
	existingConfig, err := k.client.AuthConfig.ByID(resourceID)
	if err != nil {
		return err
	}

	updateInput := map[string]interface{}{
		"enabled": false,
	}

	var result management.AuthConfig
	err = k.client.Ops.DoUpdate(resourceType, &existingConfig.Resource, updateInput, &result)
	return err
}

func (k *KeycloakClient) Update(existing *management.AuthConfig, updates interface{}) (*management.AuthConfig, error) {
	var result management.AuthConfig
	err := k.client.Ops.DoUpdate(resourceType, &existing.Resource, updates, &result)
	return &result, err
}

func (k *KeycloakClient) newEnableInputFromConfig() *management.KeyCloakConfig {
	input := &management.KeyCloakConfig{}

	input.Enabled = true
	input.AccessMode = k.Config.AccessMode
	input.Type = "keycloakConfig"

	input.RancherAPIHost = k.Config.RancherApiHost
	input.DisplayNameField = k.Config.DisplayNameField
	input.GroupsField = k.Config.GroupsField
	input.UIDField = k.Config.UIDField
	input.UserNameField = k.Config.UserNameField
	input.IDPMetadataContent = k.Config.IDPMetadataContent
	input.SpCert = k.Config.SpCert
	input.SpKey = k.Config.SpKey
	input.AllowedPrincipalIDs = k.Config.AllowedPrincipalIDs

	// input.Username = k.Config.Users.Admin.Username
	// input.Password = k.Config.Users.Admin.Password

	return input
}

// Login simulates a full browser SAML login flow
func (k *KeycloakClient) Login(username, password string) (string, error) {
	// Setup a cookie jar to hold cookies across redirects
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:       jar,
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	rancherHost := k.Config.RancherApiHost

	// Generate SAML request URL from Rancher API
	startURL := fmt.Sprintf("%s/v3-public/keycloakProviders/keycloak?action=login", rancherHost)

	// Create the JSON payload specifying the final destination after login
	payload := fmt.Sprintf(`{"finalRedirectUrl": "%s/dashboard"}`, rancherHost)

	req, err := http.NewRequest(http.MethodPost, startURL, strings.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to build login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to hit rancher login api: %w", err)
	}
	defer resp.Body.Close()

	bodyBytesCheck, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("expected 200 OK from Rancher API, got %d: %s", resp.StatusCode, string(bodyBytesCheck))
	}

	var loginActionResp struct {
		IdpRedirectUrl string `json:"idpRedirectUrl"`
	}

	if err := json.Unmarshal(bodyBytesCheck, &loginActionResp); err != nil {
		return "", fmt.Errorf("failed to parse Rancher login response: %w", err)
	}

	if loginActionResp.IdpRedirectUrl == "" {
		return "", fmt.Errorf("Rancher API did not return an idpRedirectUrl. Body: %s", string(bodyBytesCheck))
	}

	resp, err = client.Get(loginActionResp.IdpRedirectUrl)
	if err != nil {
		return "", fmt.Errorf("failed to navigate to keycloak idp url: %w", err)
	}
	defer resp.Body.Close()

	// Verify we landed on Keycloak (.../auth/realms/...)
	if !strings.Contains(resp.Request.URL.String(), "realms") {
		return "", fmt.Errorf("expected to be at keycloak login page, but valid url was: %s", resp.Request.URL)
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	//resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	bodyString := string(bodyBytes)
	actionRegex := regexp.MustCompile(`action="([^"]+)"`)
	matches := actionRegex.FindStringSubmatch(bodyString)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not find login form action in keycloak page")
	}
	loginActionURL := strings.ReplaceAll(matches[1], "&amp;", "&")

	formValues := url.Values{}
	formValues.Set("username", username)
	formValues.Set("password", password)
	formValues.Set("credentialId", "")

	// Keycloak creates a new Redirect (302) back to Rancher after this POST
	resp, err = client.PostForm(loginActionURL, formValues)
	if err != nil {
		return "", fmt.Errorf("failed to post credentials to keycloak: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ = io.ReadAll(resp.Body)
	bodyString = string(bodyBytes)
	samlRegex := regexp.MustCompile(`name="SAMLResponse" value="([^"]+)"`)
	samlMatches := samlRegex.FindStringSubmatch(bodyString)
	if len(samlMatches) < 2 {
		if strings.Contains(bodyString, "Invalid username or password") {
			return "", fmt.Errorf("login failed: invalid username or password")
		}
		return "", fmt.Errorf("could not find SAMLResponse in keycloak response")
	}
	samlResponse := samlMatches[1]
	decodedSAML, err := base64.StdEncoding.DecodeString(samlResponse)
	if err == nil {
		logrus.Debugf("decoded SAML token from Keycloak: %s", string(decodedSAML))
	}

	// Find the ACS URL (The Rancher URL to post back to)
	acsRegex := regexp.MustCompile(`action="([^"]+)"`)
	acsMatches := acsRegex.FindStringSubmatch(bodyString)
	if len(acsMatches) < 2 {
		return "", fmt.Errorf("could not find ACS URL in keycloak response")
	}
	acsURL := acsMatches[1]

	// Submit SAML response to Rancher
	rancherValues := url.Values{}
	rancherValues.Set("SAMLResponse", samlResponse)
	relayRegex := regexp.MustCompile(`name="RelayState" value="([^"]+)"`)
	relayMatches := relayRegex.FindStringSubmatch(bodyString)
	if len(relayMatches) >= 2 {
		rancherValues.Set("RelayState", relayMatches[1])
	}

	resp, err = client.PostForm(acsURL, rancherValues)
	if err != nil {
		return "", fmt.Errorf("failed to send SAMLResponse to rancher: %w", err)
	}
	defer resp.Body.Close()

	finalURL := resp.Request.URL.String()
	// If Rancher bounced us to an error page, print the HTML
	if strings.Contains(finalURL, "error") || strings.Contains(finalURL, "login") {
		errBody, _ := io.ReadAll(resp.Body)
		logrus.Warnf("keycloak login ended on URL %s with body: %s", finalURL, string(errBody))
	}

	// If successful, Rancher sets a cookie named "R_SESS" (token).
	// We can inspect the jar or the response cookies.
	// // Fallback: Sometimes the token is in the final URL fragment or body depending on settings
	// // But R_SESS is the standard for web access.
	// return "", fmt.Errorf("login completed but no R_SESS cookie found. Login might have failed silently")
	urlObj, _ := url.Parse(rancherHost)
	for _, cookie := range jar.Cookies(urlObj) {
		// Sometimes Rancher uses different casing or prefixes for the session cookie
		if strings.Contains(strings.ToUpper(cookie.Name), "SESS") {
			return cookie.Value, nil
		}
	}
	return "", fmt.Errorf("login completed but no session cookie found in jar")
}

// UpdateAccessMode safely changes the access rules without zeroing out SAML fields
func (k *KeycloakClient) UpdateAccessMode(accessMode string, allowedPrincipalIDs []string) error {
	existingConfig, err := k.client.AuthConfig.ByID(resourceID)
	if err != nil {
		return err
	}

	updateInput := k.newEnableInputFromConfig()
	updateInput.AccessMode = accessMode
	updateInput.AllowedPrincipalIDs = allowedPrincipalIDs

	var result management.AuthConfig
	return k.client.Ops.DoUpdate(resourceType, &existingConfig.Resource, updateInput, &result)
}
