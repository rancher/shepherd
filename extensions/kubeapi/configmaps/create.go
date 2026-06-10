package configmaps

import (
	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	corev1 "k8s.io/api/core/v1"
)

// CreateConfigMapWithTemplate creates a ConfigMap in the specified namespace and cluster using wrangler context.
func CreateConfigMapWithTemplate(client *rancher.Client, clusterID string, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	createdConfigMap, err := clusterContext.Core.ConfigMap().Create(configMap)
	if err != nil {
		return nil, err
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
			if err != nil {
				return err
			}

			return DeleteConfigMap(adminClient, clusterID, createdConfigMap.Namespace, createdConfigMap.Name, true)
		})
	}

	return createdConfigMap, nil
}
