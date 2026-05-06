package namespaces

import (
	"github.com/rancher/shepherd/clients/rancher"
	clusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceList is a struct that contains a list of Namespaces.
type NamespaceList struct {
	Items []corev1.Namespace
}

// ListNamespaces returns namespaces in a cluster using wrangler context and returns a NamespaceList.
func ListNamespaces(client *rancher.Client, clusterID string, listOpts metav1.ListOptions) (*NamespaceList, error) {
	clusterContext, err := clusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	namespaceList, err := clusterContext.Core.Namespace().List(listOpts)
	if err != nil {
		return nil, err
	}

	return &NamespaceList{Items: namespaceList.Items}, nil
}

// Names returns each Namespace name in the list as a new slice of strings.
func (list *NamespaceList) Names() []string {
	var names []string
	for _, ns := range list.Items {
		names = append(names, ns.Name)
	}
	return names
}
