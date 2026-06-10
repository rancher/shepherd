package secrets

import (
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	corev1 "k8s.io/api/core/v1"
)

// CreateSecretWithTemplate creates a secret using the provided template in the specified cluster using the wrangler context
func CreateSecretWithTemplate(client *rancher.Client, clusterID string, secretTemplate *corev1.Secret) (*corev1.Secret, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	createdSecret, err := clusterContext.Core.Secret().Create(secretTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret: %w", err)
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
			if err != nil {
				return err
			}

			return DeleteSecret(adminClient, clusterID, createdSecret.Namespace, createdSecret.Name, true)
		})
	}

	return createdSecret, nil
}
