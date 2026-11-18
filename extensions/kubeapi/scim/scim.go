package scim

import (
	"context"
	"fmt"
	"net/http"

	scimclient "github.com/rancher/shepherd/clients/rancher/auth/scim"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/pkg/session"
	shepherdwait "github.com/rancher/shepherd/pkg/wait"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
			delResp, delErr := scimClient.Users().Delete(id)
			if delErr != nil {
				return delErr
			}

			if delResp.StatusCode == http.StatusNotFound {
				return nil
			}

			if err := shepherdwait.ForDeletion(func(context.Context) error {
				getResp, getErr := scimClient.Users().ByID(id)
				if getErr != nil {
					return getErr
				}

				switch getResp.StatusCode {
				case http.StatusNotFound:
					return apierrors.NewNotFound(schema.GroupResource{Resource: "users"}, id)
				case http.StatusOK:
					return nil
				default:
					return fmt.Errorf("GET /Users/%s answered %d", id, getResp.StatusCode)
				}
			}, defaults.OneMinuteTimeout); err != nil {
				logrus.Warnf("SCIM user %s was deleted but %v", id, err)
			}

			return nil
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
