package keycloak

import (
	"encoding/json"
	"fmt"
)

// OIDCDiscoveryDocument represents the realm's OpenID Connect provider metadata
type OIDCDiscoveryDocument struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserInfoEndpoint      string `json:"userinfo_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
	EndSessionEndpoint    string `json:"end_session_endpoint"`
}

// SAMLDescriptor returns the realm's SAML 2.0 identity provider metadata
func (c *Client) SAMLDescriptor() (string, error) {
	descriptor, err := c.getPublic(fmt.Sprintf("/realms/%s/protocol/saml/descriptor", c.Config.Realm))
	if err != nil {
		return "", fmt.Errorf("fetching the SAML descriptor for realm %s: %w", c.Config.Realm, err)
	}

	return string(descriptor), nil
}

// OIDCDiscovery returns the realm's OpenID Connect provider metadata
func (c *Client) OIDCDiscovery() (*OIDCDiscoveryDocument, error) {
	payload, err := c.getPublic(fmt.Sprintf("/realms/%s/.well-known/openid-configuration", c.Config.Realm))
	if err != nil {
		return nil, fmt.Errorf("fetching the OpenID Connect metadata for realm %s: %w", c.Config.Realm, err)
	}

	document := new(OIDCDiscoveryDocument)
	if err := json.Unmarshal(payload, document); err != nil {
		return nil, fmt.Errorf("decoding the OpenID Connect metadata for realm %s: %w", c.Config.Realm, err)
	}

	if document.Issuer == "" || document.AuthorizationEndpoint == "" || document.TokenEndpoint == "" {
		return nil, fmt.Errorf("the OpenID Connect metadata for realm %s names no issuer, authorization endpoint or token endpoint", c.Config.Realm)
	}

	return document, nil
}
