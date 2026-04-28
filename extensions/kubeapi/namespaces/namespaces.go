package namespaces

import (
	"github.com/rancher/shepherd/clients/rancher"
	clusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetNamespaceByName is a helper function that returns the namespace by name in a specific cluster.
func GetNamespaceByName(client *rancher.Client, clusterID, namespaceName string) (*corev1.Namespace, error) {
	wranglerCtx, err := clusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	return wranglerCtx.Core.Namespace().Get(namespaceName, metav1.GetOptions{})
}
