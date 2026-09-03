package keycloak

import (
	"fmt"
	"net/http"
	"net/url"
)

// GetUser returns the realm account with the given username, or nil when there is none
func (c *Client) GetUser(username string) (*UserRepresentation, error) {
	var users []UserRepresentation

	path := c.adminPath("/users?exact=true&username=%s", url.QueryEscape(username))
	if err := c.do(http.MethodGet, path, nil, &users); err != nil {
		return nil, fmt.Errorf("looking up user %s in realm %s: %w", username, c.Config.Realm, err)
	}

	for i, existing := range users {
		if existing.Username == username {
			return &users[i], nil
		}
	}

	return nil, nil
}

// CreateUser adds an account to the realm and registers it for deletion on the session
func (c *Client) CreateUser(user *UserRepresentation) (*UserRepresentation, error) {
	if err := c.do(http.MethodPost, c.adminPath("/users"), user, nil); err != nil {
		return nil, fmt.Errorf("creating user %s in realm %s: %w", user.Username, c.Config.Realm, err)
	}

	created, err := c.GetUser(user.Username)
	if err != nil {
		return nil, err
	}

	if created == nil {
		return nil, fmt.Errorf("Keycloak accepted user %s but the realm does not have it", user.Username)
	}

	if c.Session != nil {
		c.Session.RegisterCleanupFunc(func() error {
			return c.DeleteUser(created.ID)
		})
	}

	return created, nil
}

// DeleteUser removes a realm account by its internal ID
func (c *Client) DeleteUser(id string) error {
	err := c.do(http.MethodDelete, c.adminPath("/users/%s", id), nil, nil)
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("deleting user %s from realm %s: %w", id, c.Config.Realm, err)
	}

	return nil
}

// GetUserGroups returns the groups an account belongs to directly
func (c *Client) GetUserGroups(userID string) ([]GroupRepresentation, error) {
	var groups []GroupRepresentation

	if err := c.do(http.MethodGet, c.adminPath("/users/%s/groups", userID), nil, &groups); err != nil {
		return nil, fmt.Errorf("listing the groups user %s belongs to in realm %s: %w", userID, c.Config.Realm, err)
	}

	return groups, nil
}

// AddUserToGroup makes an account a member of a group
func (c *Client) AddUserToGroup(userID, groupID string) error {
	path := c.adminPath("/users/%s/groups/%s", userID, groupID)
	if err := c.do(http.MethodPut, path, nil, nil); err != nil {
		return fmt.Errorf("adding user %s to group %s in realm %s: %w", userID, groupID, c.Config.Realm, err)
	}

	return nil
}
