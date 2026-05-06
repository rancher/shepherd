package secrets

import (
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetSecretByName retrieves a secret by name from a specific namespace in the given cluster using the wrangler context
func GetSecretByName(client *rancher.Client, clusterID, namespaceName, secretName string) (*corev1.Secret, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	secret, err := clusterContext.Core.Secret().Get(namespaceName, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret %s/%s: %w", namespaceName, secretName, err)
	}

	return secret, nil
}
