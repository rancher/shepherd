package cognito

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rancher/shepherd/extensions/defaults"
)

// OIDCDiscoveryDocument represents the user pool's OpenID Connect provider metadata
type OIDCDiscoveryDocument struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserInfoEndpoint      string `json:"userinfo_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
	EndSessionEndpoint    string `json:"end_session_endpoint"`
}

// Endpoints names the addresses Rancher drives the authorization code flow and the single logout against
type Endpoints struct {
	Issuer             string
	AuthEndpoint       string
	TokenEndpoint      string
	UserInfoEndpoint   string
	JWKSUrl            string
	EndSessionEndpoint string
}

// OIDCDiscovery returns the user pool's OpenID Connect provider metadata
func (c *Client) OIDCDiscovery(userPoolID string) (*OIDCDiscoveryDocument, error) {
	address := fmt.Sprintf("%s/.well-known/openid-configuration", c.IssuerOf(userPoolID))

	httpClient := &http.Client{
		Timeout:   defaults.OneMinuteTimeout,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}},
	}

	response, err := httpClient.Get(address)
	if err != nil {
		return nil, fmt.Errorf("fetching the OpenID Connect metadata for user pool %s: %w", userPoolID, err)
	}
	defer response.Body.Close()

	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading the OpenID Connect metadata for user pool %s: %w", userPoolID, err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Cognito answered %d for the OpenID Connect metadata of user pool %s: %s",
			response.StatusCode, userPoolID, payload)
	}

	document := new(OIDCDiscoveryDocument)
	if err := json.Unmarshal(payload, document); err != nil {
		return nil, fmt.Errorf("decoding the OpenID Connect metadata for user pool %s: %w", userPoolID, err)
	}

	if document.Issuer == "" || document.JWKSURI == "" {
		return nil, fmt.Errorf("the OpenID Connect metadata for user pool %s names no issuer or JWKS address",
			userPoolID)
	}

	return document, nil
}

// EndpointsFor returns the addresses the pool serves
func (c *Client) EndpointsFor(userPoolID, domain string) (*Endpoints, error) {
	if userPoolID == "" {
		return nil, fmt.Errorf("a Cognito user pool identifier is needed to resolve the addresses Rancher " +
			"drives the authorization code flow against")
	}

	document, err := c.OIDCDiscovery(userPoolID)
	if err != nil {
		return nil, err
	}

	endpoints := &Endpoints{
		Issuer:             document.Issuer,
		AuthEndpoint:       document.AuthorizationEndpoint,
		TokenEndpoint:      document.TokenEndpoint,
		UserInfoEndpoint:   document.UserInfoEndpoint,
		JWKSUrl:            document.JWKSURI,
		EndSessionEndpoint: document.EndSessionEndpoint,
	}

	if domain != "" {
		if endpoints.AuthEndpoint == "" {
			endpoints.AuthEndpoint = domain + "/oauth2/authorize"
		}

		if endpoints.TokenEndpoint == "" {
			endpoints.TokenEndpoint = domain + "/oauth2/token"
		}

		if endpoints.UserInfoEndpoint == "" {
			endpoints.UserInfoEndpoint = domain + "/oauth2/userInfo"
		}

		if endpoints.EndSessionEndpoint == "" {
			endpoints.EndSessionEndpoint = domain + "/logout"
		}
	}

	if endpoints.AuthEndpoint == "" || endpoints.TokenEndpoint == "" {
		return nil, fmt.Errorf("neither discovery nor a hosted UI domain names the authorization and token "+
			"endpoints of user pool %s, so there is no code flow to drive", userPoolID)
	}

	return endpoints, nil
}
