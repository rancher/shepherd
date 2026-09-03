package keycloak

import (
	"fmt"
	"net/http"
	"net/url"
)

// GetClient returns the realm client registered under the given client ID, or nil when there is none
func (c *Client) GetClient(clientID string) (*ClientRepresentation, error) {
	var clients []ClientRepresentation

	path := c.adminPath("/clients?clientId=%s", url.QueryEscape(clientID))
	if err := c.do(http.MethodGet, path, nil, &clients); err != nil {
		return nil, fmt.Errorf("looking up the %s client in realm %s: %w", clientID, c.Config.Realm, err)
	}

	for i, existing := range clients {
		if existing.ClientID == clientID {
			return &clients[i], nil
		}
	}

	return nil, nil
}

// CreateClient registers a client in the realm and registers it for deletion on the session
func (c *Client) CreateClient(client *ClientRepresentation) (*ClientRepresentation, error) {
	if err := c.do(http.MethodPost, c.adminPath("/clients"), client, nil); err != nil {
		return nil, fmt.Errorf("creating the %s client in realm %s: %w", client.ClientID, c.Config.Realm, err)
	}

	created, err := c.GetClient(client.ClientID)
	if err != nil {
		return nil, err
	}

	if created == nil {
		return nil, fmt.Errorf("Keycloak accepted the %s client but the realm does not have it", client.ClientID)
	}

	if c.Session != nil {
		c.Session.RegisterCleanupFunc(func() error {
			return c.DeleteClient(created.ID)
		})
	}

	return created, nil
}

// DeleteClient removes a realm client by its internal ID
func (c *Client) DeleteClient(id string) error {
	err := c.do(http.MethodDelete, c.adminPath("/clients/%s", id), nil, nil)
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("deleting client %s from realm %s: %w", id, c.Config.Realm, err)
	}

	return nil
}

// ReplaceClient registers a client, first removing any client already holding the same client ID
func (c *Client) ReplaceClient(client *ClientRepresentation) (*ClientRepresentation, error) {
	existing, err := c.GetClient(client.ClientID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		if err := c.DeleteClient(existing.ID); err != nil {
			return nil, err
		}
	}

	return c.CreateClient(client)
}

// GetProtocolMapper returns the mapper of the given name on a realm client, or nil when there is none
func (c *Client) GetProtocolMapper(clientUUID, name string) (*ProtocolMapperRepresentation, error) {
	var mappers []ProtocolMapperRepresentation

	path := c.adminPath("/clients/%s/protocol-mappers/models", clientUUID)
	if err := c.do(http.MethodGet, path, nil, &mappers); err != nil {
		return nil, fmt.Errorf("listing the mappers on client %s: %w", clientUUID, err)
	}

	for i, mapper := range mappers {
		if mapper.Name == name {
			return &mappers[i], nil
		}
	}

	return nil, nil
}

// ReplaceProtocolMapper writes an existing mapper back with the given configuration
func (c *Client) ReplaceProtocolMapper(clientUUID string, mapper *ProtocolMapperRepresentation) error {
	if mapper.ID == "" {
		return fmt.Errorf("the %s mapper on client %s cannot be replaced without the ID Keycloak stored it under", mapper.Name, clientUUID)
	}

	path := c.adminPath("/clients/%s/protocol-mappers/models/%s", clientUUID, mapper.ID)
	if err := c.do(http.MethodPut, path, mapper, nil); err != nil {
		return fmt.Errorf("replacing the %s mapper on client %s: %w", mapper.Name, clientUUID, err)
	}

	return nil
}
