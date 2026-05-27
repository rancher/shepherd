package secrets

import (
	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretList is a struct that contains a list of Secrets.
type SecretList struct {
	Items []corev1.Secret
}

// ListSecrets is a helper function that uses the wrangler context to list secrets in a cluster and returns a SecretList.
func ListSecrets(client *rancher.Client, clusterID, namespace string, listOpts metav1.ListOptions) (*SecretList, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	secrets, err := clusterContext.Core.Secret().List(namespace, listOpts)
	if err != nil {
		return nil, err
	}

	return &SecretList{Items: secrets.Items}, nil
}

// Names returns each Secret name in the list as a new slice of strings.
func (list *SecretList) Names() []string {
	var names []string
	for _, s := range list.Items {
		names = append(names, s.Name)
	}
	return names
}
