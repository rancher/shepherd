package keycloak

import (
	"fmt"
	"net/http"
)

// GetRealm returns the administered realm, or nil when the server has no such realm
func (c *Client) GetRealm() (*RealmRepresentation, error) {
	var realm RealmRepresentation

	err := c.do(http.MethodGet, fmt.Sprintf("/admin/realms/%s", c.Config.Realm), nil, &realm)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return &realm, nil
}

// CreateRealm adds the administered realm to the server
func (c *Client) CreateRealm() error {
	realm := &RealmRepresentation{
		Realm:   c.Config.Realm,
		Enabled: Pointer(true),
	}

	if err := c.do(http.MethodPost, "/admin/realms", realm, nil); err != nil {
		return fmt.Errorf("creating the %s realm: %w", c.Config.Realm, err)
	}

	return nil
}

// DeleteRealm removes the administered realm from the server
func (c *Client) DeleteRealm() error {
	err := c.do(http.MethodDelete, fmt.Sprintf("/admin/realms/%s", c.Config.Realm), nil, nil)
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("deleting the %s realm: %w", c.Config.Realm, err)
	}

	return nil
}

// EnsureRealm creates the administered realm when the server does not already have it, and reports whether it had to
func (c *Client) EnsureRealm() (created bool, err error) {
	existing, err := c.GetRealm()
	if err != nil {
		return false, err
	}

	if existing != nil {
		return false, nil
	}

	if err := c.CreateRealm(); err != nil {
		return false, err
	}

	if c.Session != nil {
		c.Session.RegisterCleanupFunc(c.DeleteRealm)
	}

	return true, nil
}
