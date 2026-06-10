package namespaces

import (
	"github.com/rancher/shepherd/clients/rancher"
	clusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	corev1 "k8s.io/api/core/v1"
)

// CreateNamespace creates a namespace in a cluster using wrangler context.
func CreateNamespace(client *rancher.Client, clusterID string, namespace *corev1.Namespace) (*corev1.Namespace, error) {
	wranglerCtx, err := clusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	createdNamespace, err := wranglerCtx.Core.Namespace().Create(namespace)
	if err != nil {
		return nil, err
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
			if err != nil {
				return err
			}

			return DeleteNamespace(adminClient, clusterID, createdNamespace.Name, true)
		})
	}

	return createdNamespace, nil
}
