package keycloak

import (
	"fmt"
	"net/http"
	"net/url"
)

// GetGroup returns the top level realm group with the given name, or nil when there is none
func (c *Client) GetGroup(name string) (*GroupRepresentation, error) {
	var groups []GroupRepresentation

	path := c.adminPath("/groups?exact=true&search=%s", url.QueryEscape(name))
	if err := c.do(http.MethodGet, path, nil, &groups); err != nil {
		return nil, fmt.Errorf("looking up group %s in realm %s: %w", name, c.Config.Realm, err)
	}

	for i, existing := range groups {
		if existing.Name == name {
			return &groups[i], nil
		}
	}

	return nil, nil
}

// CreateGroup adds a top level group to the realm and registers it for deletion on the session
func (c *Client) CreateGroup(name string) (*GroupRepresentation, error) {
	group := &GroupRepresentation{Name: name}

	if err := c.do(http.MethodPost, c.adminPath("/groups"), group, nil); err != nil {
		return nil, fmt.Errorf("creating group %s in realm %s: %w", name, c.Config.Realm, err)
	}

	created, err := c.GetGroup(name)
	if err != nil {
		return nil, err
	}

	if created == nil {
		return nil, fmt.Errorf("Keycloak accepted group %s but the realm does not have it", name)
	}

	if c.Session != nil {
		c.Session.RegisterCleanupFunc(func() error {
			return c.DeleteGroup(created.ID)
		})
	}

	return created, nil
}

// GetChildGroup returns the group of the given name directly beneath the given parent, or nil when there is none
func (c *Client) GetChildGroup(parentID, name string) (*GroupRepresentation, error) {
	var children []GroupRepresentation

	err := c.do(http.MethodGet, c.adminPath("/groups/%s/children", parentID), nil, &children)
	if err != nil {
		if !isNotFound(err) {
			return nil, fmt.Errorf("listing the children of group %s in realm %s: %w", parentID, c.Config.Realm, err)
		}

		parent := &GroupRepresentation{}
		if err := c.do(http.MethodGet, c.adminPath("/groups/%s", parentID), nil, parent); err != nil {
			return nil, fmt.Errorf("reading group %s in realm %s: %w", parentID, c.Config.Realm, err)
		}

		children = parent.SubGroups
	}

	for i, child := range children {
		if child.Name == name {
			return &children[i], nil
		}
	}

	return nil, nil
}

// CreateChildGroup adds a group beneath an existing one and registers it for deletion on the session
func (c *Client) CreateChildGroup(parentID, name string) (*GroupRepresentation, error) {
	created := &GroupRepresentation{}

	path := c.adminPath("/groups/%s/children", parentID)
	if err := c.do(http.MethodPost, path, &GroupRepresentation{Name: name}, created); err != nil {
		return nil, fmt.Errorf("creating group %s beneath group %s in realm %s: %w", name, parentID, c.Config.Realm, err)
	}

	if created.ID == "" {
		found, err := c.GetChildGroup(parentID, name)
		if err != nil {
			return nil, err
		}

		if found == nil {
			return nil, fmt.Errorf("Keycloak accepted group %s beneath group %s but the realm does not have it", name, parentID)
		}

		created = found
	}

	if c.Session != nil {
		c.Session.RegisterCleanupFunc(func() error {
			return c.DeleteGroup(created.ID)
		})
	}

	return created, nil
}

// DeleteGroup removes a realm group by its internal ID
func (c *Client) DeleteGroup(id string) error {
	err := c.do(http.MethodDelete, c.adminPath("/groups/%s", id), nil, nil)
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("deleting group %s from realm %s: %w", id, c.Config.Realm, err)
	}

	return nil
}
