package oidcclient

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	oidcext "github.com/rancher/shepherd/extensions/auth/oidc"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

const OIDCClientSecretNamespace = "cattle-oidc-client-secrets"

var oidcClientGVR = schema.GroupVersionResource{
	Group:    oidcext.OIDCClientGroup,
	Version:  oidcext.OIDCClientVersion,
	Resource: oidcext.OIDCClientResource,
}

type ClientSpec struct {
	RedirectURIs                  []string
	Scopes                        []string
	TokenExpirationSeconds        int
	RefreshTokenExpirationSeconds int
}

// CreateOIDCClient creates an OIDCClient CRD and registers DeleteOIDCClient as session cleanup.
func CreateOIDCClient(client *rancher.Client, name string, spec ClientSpec) error {
	logrus.Infof("[OIDC setup] Creating OIDCClient CRD %q on management cluster", name)
	redirectURIs := make([]interface{}, len(spec.RedirectURIs))
	for i, v := range spec.RedirectURIs {
		redirectURIs[i] = v
	}
	scopes := make([]interface{}, len(spec.Scopes))
	for i, v := range spec.Scopes {
		scopes[i] = v
	}
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": oidcext.OIDCClientGroup + "/" + oidcext.OIDCClientVersion,
			"kind":       oidcext.OIDCClientKind,
			"metadata": map[string]interface{}{
				"name": name,
			},
			"spec": map[string]interface{}{
				"redirectURIs":                  redirectURIs,
				"scopes":                        scopes,
				"tokenExpirationSeconds":        int64(spec.TokenExpirationSeconds),
				"refreshTokenExpirationSeconds": int64(spec.RefreshTokenExpirationSeconds),
			},
		},
	}
	dynClient, err := client.GetRancherDynamicClient()
	if err != nil {
		return err
	}
	_, err = dynClient.Resource(oidcClientGVR).Create(context.Background(), obj, metav1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return fmt.Errorf("creating OIDCClient %q: %w", name, err)
		}
		logrus.Infof("[OIDC setup] OIDCClient %q already exists — skipping creation", name)
		return nil
	}
	logrus.Infof("[OIDC setup] OIDCClient %q created", name)
	client.Session.RegisterCleanupFunc(func() error {
		return DeleteOIDCClient(client, name)
	})
	return nil
}

// WaitForOIDCClientReady polls until status.clientID and status.clientSecrets are populated.
func WaitForOIDCClientReady(client *rancher.Client, name string) (clientID, secretKeyName string, err error) {
	logrus.Infof("[OIDC setup] Waiting for OIDCClient %q status.clientID (max 2m)", name)
	dynClient, err := client.GetRancherDynamicClient()
	if err != nil {
		return "", "", err
	}
	err = kwait.PollUntilContextTimeout(
		context.Background(), defaults.FiveSecondTimeout, defaults.TwoMinuteTimeout, true,
		func(ctx context.Context) (bool, error) {
			obj, getErr := dynClient.Resource(oidcClientGVR).Get(ctx, name, metav1.GetOptions{})
			if getErr != nil {
				logrus.Debugf("[OIDC] OIDCClient %q not yet visible: %v", name, getErr)
				return false, nil
			}
			status, ok := obj.Object["status"].(map[string]interface{})
			if !ok {
				return false, nil
			}
			id, _ := status["clientID"].(string)
			if id == "" {
				return false, nil
			}
			secrets, _ := status["clientSecrets"].(map[string]interface{})
			if len(secrets) == 0 {
				return false, nil
			}
			for k := range secrets {
				secretKeyName = k
				break
			}
			clientID = id
			logrus.Infof("[OIDC] OIDCClient %q ready — clientID=%s secretKey=%s", name, clientID, secretKeyName)
			return true, nil
		},
	)
	if err != nil {
		return "", "", fmt.Errorf("timed out waiting for OIDCClient %q status.clientID: %w", name, err)
	}
	return clientID, secretKeyName, nil
}

// FetchOIDCClientSecret retrieves the OIDCClient secret value from the cattle-oidc-client-secrets namespace.
func FetchOIDCClientSecret(client *rancher.Client, clientID, secretKeyName string) (string, error) {
	logrus.Infof("[OIDC setup] Fetching client secret from %s/%s key=%s",
		OIDCClientSecretNamespace, clientID, secretKeyName)
	secret, err := client.WranglerContext.Core.Secret().Get(
		OIDCClientSecretNamespace, clientID, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("getting OIDCClient secret %s/%s: %w",
			OIDCClientSecretNamespace, clientID, err)
	}
	value, ok := secret.Data[secretKeyName]
	if !ok || len(value) == 0 {
		return "", fmt.Errorf("key %q not found or empty in secret %s/%s",
			secretKeyName, OIDCClientSecretNamespace, clientID)
	}
	return string(value), nil
}

// DeleteOIDCClient deletes the OIDCClient by name; NotFound is treated as success.
func DeleteOIDCClient(client *rancher.Client, name string) error {
	logrus.Infof("[OIDC teardown] Deleting OIDCClient %q", name)
	dynClient, err := client.GetRancherDynamicClient()
	if err != nil {
		return err
	}
	err = dynClient.Resource(oidcClientGVR).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logrus.Debugf("[OIDC teardown] OIDCClient %q already gone — skipping", name)
			return nil
		}
		return fmt.Errorf("deleting OIDCClient %q: %w", name, err)
	}
	logrus.Infof("[OIDC teardown] OIDCClient %q deleted", name)
	return nil
}
