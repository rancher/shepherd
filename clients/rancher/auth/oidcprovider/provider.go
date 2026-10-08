package oidcprovider

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
	oidcext "github.com/rancher/shepherd/extensions/auth/oidcprovider"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/pkg/clientbase"
	"github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
)

const (
	managementAPIPath = "/v3"
	publicAPIPath     = "/v3-public"
	verifyAuthPath    = "/verify-auth"
	redirectToIdPPath = "/v1-oidc"

	stateParam = "state"
	scopeParam = "scope"

	configureTestAction = "configureTest"
	testAndApplyAction  = "testAndApply"
	loginAction         = "login"

	jsonResponseType = "json"
	openIDScope      = "openid"
	stateBytes       = 12
)

type ProviderClient struct {
	client     *management.Client
	session    *session.Session
	provider   Provider
	rancherURL string
	transport  *http.Transport
	httpClient *http.Client

	Config *Config
}

// NewProviderClient constructs an OpenID Connect provider struct after it reads the provider's configuration key
func NewProviderClient(client *management.Client, session *session.Session, provider Provider) (*ProviderClient, error) {
	if provider.Name == "" || provider.ConfigType == "" || provider.PublicType == "" || provider.ConfigKey == "" {
		return nil, fmt.Errorf("an OpenID Connect provider needs a name, a config type, a public type and a config key, got %+v", provider)
	}

	providerConfig := new(Config)
	config.LoadConfig(provider.ConfigKey, providerConfig)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint:gosec
	}

	return &ProviderClient{
		client:     client,
		session:    session,
		provider:   provider,
		rancherURL: strings.TrimSuffix(client.Opts.URL, managementAPIPath),
		transport:  transport,
		httpClient: &http.Client{Transport: transport, Timeout: defaults.OneMinuteTimeout},
		Config:     providerConfig,
	}, nil
}

// Provider returns the provider this client drives
func (p *ProviderClient) Provider() Provider {
	return p.provider
}

// redirectURL names the Rancher address the identity provider sends the authorization code back to
func (p *ProviderClient) redirectURL() string {
	if p.Config.RancherURL != "" {
		return p.Config.RancherURL
	}

	return p.rancherURL + verifyAuthPath
}

// TrustIdentityProvider makes Rancher accept the certificate the identity provider serves, by bundling it behind a generated client key pair
func (p *ProviderClient) TrustIdentityProvider() error {
	if p.Config.Issuer == "" {
		return fmt.Errorf("issuer is empty under the %s config key, so there is no address to read the "+
			"identity provider's certificate from", p.provider.ConfigKey)
	}

	address, err := oidcext.TLSAddressOf(p.Config.Issuer)
	if err != nil {
		return err
	}

	served, err := oidcext.FetchServerCertificates(address)
	if err != nil {
		return err
	}

	clientKeyPair, err := newClientKeyPair(p.provider.Name)
	if err != nil {
		return err
	}

	p.Config.Certificate = clientKeyPair.Certificate + served
	p.Config.PrivateKey = clientKeyPair.PrivateKey

	return nil
}

// ConfigureTest asks Rancher to build the authorization request that proves the configuration works
func (p *ProviderClient) ConfigureTest() (string, error) {
	input, err := p.newConfigInputWithSecrets()
	if err != nil {
		return "", err
	}

	var output management.OIDCTestOutput
	if err := p.client.Ops.DoModify(http.MethodPost, p.actionURL(configureTestAction), input, &output); err != nil {
		return "", fmt.Errorf("asking Rancher to build the %s authorization request: %w", p.provider.Name, err)
	}

	if output.RedirectURL == "" {
		return "", fmt.Errorf("Rancher accepted the %s configuration but named no address to send the "+
			"administrator to, so there is no sign-in to complete", p.provider.Name)
	}

	return output.RedirectURL, nil
}

// EnableWithAdminLogin turns the provider on by having an administrator sign in at the identity provider
func (p *ProviderClient) EnableWithAdminLogin(username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("enabling %s needs the credentials of an identity provider account for the "+
			"administrator, set users.admin under the %s config key", p.provider.Name, p.provider.ConfigKey)
	}

	authorizationURL, err := p.authorizationURL()
	if err != nil {
		return err
	}

	code, err := oidcext.CaptureAuthorizationCode(p.transport, authorizationURL, p.redirectURL(), username, password)
	if err != nil {
		return fmt.Errorf("signing the administrator in at the %s identity provider: %w", p.provider.Name, err)
	}

	input, err := p.newConfigInputWithSecrets()
	if err != nil {
		return err
	}

	input["enabled"] = true

	applyInput := map[string]any{
		"oidcConfig": input,
		"code":       code,
	}

	var result management.AuthConfig
	if err := p.client.Ops.DoModify(http.MethodPost, p.actionURL(testAndApplyAction), applyInput, &result); err != nil {
		return fmt.Errorf("applying the %s configuration with the administrator's authorization code: %w",
			p.provider.Name, err)
	}

	p.session.RegisterCleanupFunc(p.Disable)

	enabled, err := p.client.AuthConfig.ByID(p.provider.Name)
	if err != nil {
		return fmt.Errorf("retrieving the %s auth config after enabling it: %w", p.provider.Name, err)
	}

	if !enabled.Enabled {
		return fmt.Errorf("Rancher accepted the authorization code for %s but the auth config is still disabled",
			p.provider.Name)
	}

	return nil
}

// Enable writes the auth config with the given configuration values, without any sign-in at the identity provider
func (p *ProviderClient) Enable() error {
	input, err := p.newConfigInputWithSecrets()
	if err != nil {
		return err
	}

	input["enabled"] = true

	if err := p.writeConfig(input); err != nil {
		return fmt.Errorf("enabling the %s auth provider: %w", p.provider.Name, err)
	}

	p.session.RegisterCleanupFunc(p.Disable)

	return nil
}

// Disable makes a request to disable the provider
func (p *ProviderClient) Disable() error {
	if err := p.writeConfig(map[string]any{
		"type":    p.provider.ConfigType,
		"enabled": false,
	}); err != nil {
		return fmt.Errorf("disabling the %s auth provider: %w", p.provider.Name, err)
	}

	return nil
}

// UpdateAccessMode changes who may sign in through the provider
func (p *ProviderClient) UpdateAccessMode(accessMode string, allowedPrincipalIDs []string) error {
	input, err := p.newConfigInput()
	if err != nil {
		return err
	}

	input["enabled"] = true
	input["accessMode"] = accessMode
	input["allowedPrincipalIds"] = allowedPrincipalIDs

	if err := p.writeConfig(input); err != nil {
		return fmt.Errorf("setting the %s access mode to %s: %w", p.provider.Name, accessMode, err)
	}

	return nil
}

// UpdateGroupsClaim points the provider at the token claim it should read group membership from
func (p *ProviderClient) UpdateGroupsClaim(groupsClaim string) error {
	input, err := p.newConfigInput()
	if err != nil {
		return err
	}

	input["enabled"] = true
	input["groupsClaim"] = groupsClaim

	if err := p.writeConfig(input); err != nil {
		return fmt.Errorf("setting the %s groups claim to %s: %w", p.provider.Name, groupsClaim, err)
	}

	p.Config.GroupsClaim = groupsClaim

	return nil
}

// authorizationURL builds the request that asks the identity provider for an authorization code
func (p *ProviderClient) authorizationURL() (string, error) {
	if p.Config.PKCEMethod != "" {
		return "", fmt.Errorf("pkceMethod is set to %s under the %s config key, and Rancher keeps the verifier "+
			"for that exchange in a browser cookie this client never holds, so leave it empty to sign in headlessly",
			p.Config.PKCEMethod, p.provider.ConfigKey)
	}

	if p.Config.AuthEndpoint == "" {
		return "", fmt.Errorf("authEndpoint is empty under the %s config key, so there is no address to ask the "+
			"identity provider for an authorization code at", p.provider.ConfigKey)
	}

	state, err := randomState()
	if err != nil {
		return "", err
	}

	return oidcext.AuthorizationURL(p.Config.AuthEndpoint, p.Config.ClientID, p.redirectURL(), p.scopes(), state)
}

// LoginAsUser signs a user in through the provider and returns the Rancher token the login issued
func (p *ProviderClient) LoginAsUser(username, password string) (*management.Token, error) {
	state, err := randomState()
	if err != nil {
		return nil, err
	}

	parameters := url.Values{}
	parameters.Set(stateParam, state)
	parameters.Set(scopeParam, p.scopes())

	startURL := fmt.Sprintf("%s%s/%s?%s", p.rancherURL, redirectToIdPPath, p.provider.Name, parameters.Encode())

	code, err := oidcext.CaptureAuthorizationCode(p.transport, startURL, p.redirectURL(), username, password)
	if err != nil {
		return nil, fmt.Errorf("signing %s in at the %s identity provider: %w", username, p.provider.Name, err)
	}

	loginURL := fmt.Sprintf("%s%s/%ss/%s?action=%s", p.rancherURL, publicAPIPath, p.provider.PublicType,
		p.provider.Name, loginAction)

	response, err := clientbase.Do(p.httpClient, http.MethodPost, loginURL, map[string]any{
		"code":         code,
		"responseType": jsonResponseType,
	}, map[string]string{"Content-Type": "application/json"})
	if err != nil {
		return nil, fmt.Errorf("presenting the authorization code for %s to Rancher: %w", username, err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("Rancher refused the %s login for %s, it answered %d: %s",
			p.provider.Name, username, response.StatusCode, response.Body)
	}

	token := new(management.Token)
	if err := json.Unmarshal(response.Body, token); err != nil {
		return nil, fmt.Errorf("decoding the token Rancher issued for %s: %w", username, err)
	}

	if token.Token == "" {
		return nil, fmt.Errorf("Rancher accepted the %s login for %s but issued no token", p.provider.Name, username)
	}

	return token, nil
}

// IDTokenClaims returns the claims the identity provider issues for a user, which is what Rancher reads an identity from
func (p *ProviderClient) IDTokenClaims(username, password string) (map[string]interface{}, error) {
	if p.Config.TokenEndpoint == "" {
		return nil, fmt.Errorf("tokenEndpoint is empty under the %s config key, so an authorization code cannot "+
			"be exchanged for the claims the identity provider asserts", p.provider.ConfigKey)
	}

	authorizationURL, err := p.authorizationURL()
	if err != nil {
		return nil, err
	}

	return oidcext.IDTokenClaims(p.transport, authorizationURL, p.Config.TokenEndpoint, p.Config.ClientID,
		p.Config.ClientSecret, p.redirectURL(), username, password)
}

// writeConfig puts the given values onto the provider's auth config
func (p *ProviderClient) writeConfig(updates map[string]any) error {
	existing, err := p.client.AuthConfig.ByID(p.provider.Name)
	if err != nil {
		return fmt.Errorf("retrieving the %s auth config: %w", p.provider.Name, err)
	}

	var result management.AuthConfig

	return p.client.Ops.DoUpdate(p.provider.ConfigType, &existing.Resource, updates, &result)
}

// actionURL names the auth config endpoint the given action is posted to
func (p *ProviderClient) actionURL(action string) string {
	return fmt.Sprintf("%s/%ss/%s?action=%s", p.client.Opts.URL, p.provider.ConfigType, p.provider.Name, action)
}

// scopes returns the scopes the authorization request asks for, falling back to the provider default
func (p *ProviderClient) scopes() string {
	scopes := strings.Fields(p.Config.Scopes)
	for _, scope := range scopes {
		if scope == openIDScope {
			return strings.Join(scopes, " ")
		}
	}

	return strings.Join(append([]string{openIDScope}, scopes...), " ")
}

// groupsClaim returns the token claim group membership is read from, falling back to the provider default
func (p *ProviderClient) groupsClaim() string {
	if p.Config.GroupsClaim == "" {
		return "groups"
	}

	return p.Config.GroupsClaim
}

// newConfigInput builds the provider config body, leaving out the secrets Rancher already holds a reference to
func (p *ProviderClient) newConfigInput() (map[string]any, error) {
	if p.Config.Issuer == "" {
		return nil, fmt.Errorf("issuer is empty under the %s config key, Rancher reads the identity provider's "+
			"endpoints from it so it must be set", p.provider.ConfigKey)
	}

	if p.Config.ClientID == "" || p.Config.ClientSecret == "" {
		return nil, fmt.Errorf("clientId and clientSecret must both be set under the %s config key, Rancher "+
			"identifies itself to the identity provider with them", p.provider.ConfigKey)
	}

	groupSearchEnabled := false
	if p.Config.GroupSearchEnabled != nil {
		groupSearchEnabled = *p.Config.GroupSearchEnabled
	}

	return map[string]any{
		"type":                      p.provider.ConfigType,
		"accessMode":                p.Config.AccessMode,
		"allowedPrincipalIds":       p.Config.AllowedPrincipalIDs,
		"clientId":                  p.Config.ClientID,
		"issuer":                    p.Config.Issuer,
		"authEndpoint":              p.Config.AuthEndpoint,
		"tokenEndpoint":             p.Config.TokenEndpoint,
		"userInfoEndpoint":          p.Config.UserInfoEndpoint,
		"jwksUrl":                   p.Config.JWKSUrl,
		"endSessionEndpoint":        p.Config.EndSessionEndpoint,
		"rancherUrl":                p.redirectURL(),
		"certificate":               p.Config.Certificate,
		"scope":                     p.scopes(),
		"groupsClaim":               p.groupsClaim(),
		"pkceMethod":                p.Config.PKCEMethod,
		"groupSearchEnabled":        groupSearchEnabled,
		"clientAuthenticatedSearch": p.Config.ClientAuthenticatedSearch,
	}, nil
}

// newConfigInputWithSecrets builds the provider config body including the client secret and private key
func (p *ProviderClient) newConfigInputWithSecrets() (map[string]any, error) {
	input, err := p.newConfigInput()
	if err != nil {
		return nil, err
	}

	input["clientSecret"] = p.Config.ClientSecret
	input["privateKey"] = p.Config.PrivateKey

	return input, nil
}

// randomState returns the opaque value the authorization request is correlated by
func randomState() (string, error) {
	raw := make([]byte, stateBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generating the state that ties an authorization request to its answer: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}
