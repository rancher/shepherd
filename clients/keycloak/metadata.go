package keycloak

import "fmt"

// SAMLDescriptor returns the realm's SAML 2.0 identity provider metadata
func (c *Client) SAMLDescriptor() (string, error) {
	descriptor, err := c.getPublic(fmt.Sprintf("/realms/%s/protocol/saml/descriptor", c.Config.Realm))
	if err != nil {
		return "", fmt.Errorf("fetching the SAML descriptor for realm %s: %w", c.Config.Realm, err)
	}

	return string(descriptor), nil
}
