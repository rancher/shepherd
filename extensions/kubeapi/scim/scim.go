package scim

import (
	"fmt"
	"net/http"

	scimclient "github.com/rancher/shepherd/clients/rancher/auth/scim"
	"github.com/rancher/shepherd/pkg/session"
)

// CreateUserWithTemplate creates a SCIM User from the given template and, on success,
// registers a session cleanup that deletes the user when the session ends.
// Returns the raw SCIM response so the caller can decode the body or assert on status.
func CreateUserWithTemplate(scimClient *scimclient.Client, ts *session.Session, user scimclient.User) (*scimclient.Response, error) {
	resp, err := scimClient.Users().Create(user)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusCreated {
		return resp, fmt.Errorf("create user: expected 201, got %d: %s", resp.StatusCode, string(resp.Body))
	}

	if ts != nil {
		id, idErr := resp.IDFromBody()
		if idErr != nil {
			return resp, fmt.Errorf("create user: could not parse id for cleanup: %w", idErr)
		}
		ts.RegisterCleanupFunc(func() error {
			_, delErr := scimClient.Users().Delete(id)
			return delErr
		})
	}

	return resp, nil
}

// CreateGroupWithTemplate creates a SCIM Group from the given template and, on success,
// registers a session cleanup that deletes the group when the session ends.
// Returns the raw SCIM response so the caller can decode the body or assert on status.
func CreateGroupWithTemplate(scimClient *scimclient.Client, ts *session.Session, group scimclient.Group) (*scimclient.Response, error) {
	resp, err := scimClient.Groups().Create(group)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusCreated {
		return resp, fmt.Errorf("create group: expected 201, got %d: %s", resp.StatusCode, string(resp.Body))
	}

	if ts != nil {
		id, idErr := resp.IDFromBody()
		if idErr != nil {
			return resp, fmt.Errorf("create group: could not parse id for cleanup: %w", idErr)
		}
		ts.RegisterCleanupFunc(func() error {
			_, delErr := scimClient.Groups().Delete(id)
			return delErr
		})
	}

	return resp, nil
}
