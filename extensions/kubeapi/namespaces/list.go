package namespaces

import (
	"github.com/rancher/shepherd/clients/rancher"
	clusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListNamespaces returns namespaces in a cluster using wrangler context.
func ListNamespaces(client *rancher.Client, clusterID string, listOpts metav1.ListOptions) (*corev1.NamespaceList, error) {
	wranglerCtx, err := clusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	namespaceList, err := wranglerCtx.Core.Namespace().List(listOpts)
	if err != nil {
		return nil, err
	}

	return namespaceList, nil
}
