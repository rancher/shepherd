package oidcclient

import (
	"context"
	"fmt"

	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

const OIDCClientSecretNamespace = "cattle-oidc-client-secrets"

// CreateOIDCClient creates an OIDCClient and registers DeleteOIDCClient as session cleanup.
func CreateOIDCClient(client *rancher.Client, name string, spec v3.OIDCClientSpec) (*v3.OIDCClient, error) {
	oidcClient := &v3.OIDCClient{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: spec,
	}

	createdClient, err := client.WranglerContext.Mgmt.OIDCClient().Create(oidcClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDCClient %s: %w", name, err)
	}

	client.Session.RegisterCleanupFunc(func() error {
		return DeleteOIDCClient(client, name)
	})

	return createdClient, nil
}

// WaitForOIDCClientReady polls until the OIDCClient reports a client ID and at least one client
// secret, returning the client ID and the name of the first secret key.
func WaitForOIDCClientReady(client *rancher.Client, name string) (string, string, error) {
	var clientID, secretKeyName string

	err := kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.TwoMinuteTimeout, true, func(ctx context.Context) (bool, error) {
		oidcClient, err := client.WranglerContext.Mgmt.OIDCClient().Get(name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return false, nil
			}

			return false, err
		}

		if oidcClient.Status.ClientID == "" || len(oidcClient.Status.ClientSecrets) == 0 {
			return false, nil
		}

		clientID = oidcClient.Status.ClientID
		for key := range oidcClient.Status.ClientSecrets {
			secretKeyName = key
			break
		}

		return true, nil
	})
	if err != nil {
		return "", "", fmt.Errorf("timed out waiting for OIDCClient %s to report a client ID and secret: %w", name, err)
	}

	logrus.Infof("OIDCClient %s is ready with client ID %s", name, clientID)

	return clientID, secretKeyName, nil
}

// FetchOIDCClientSecret returns the OIDCClient secret value stored under the given key.
func FetchOIDCClientSecret(client *rancher.Client, clientID, secretKeyName string) (string, error) {
	secret, err := client.WranglerContext.Core.Secret().Get(OIDCClientSecretNamespace, clientID, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get OIDCClient secret %s/%s: %w", OIDCClientSecretNamespace, clientID, err)
	}

	value, ok := secret.Data[secretKeyName]
	if !ok || len(value) == 0 {
		return "", fmt.Errorf("key %s not found or empty in secret %s/%s", secretKeyName, OIDCClientSecretNamespace, clientID)
	}

	return string(value), nil
}

// DeleteOIDCClient deletes the OIDCClient with the given name, treating NotFound as success.
func DeleteOIDCClient(client *rancher.Client, name string) error {
	err := client.WranglerContext.Mgmt.OIDCClient().Delete(name, &metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete OIDCClient %s: %w", name, err)
	}

	return nil
}
