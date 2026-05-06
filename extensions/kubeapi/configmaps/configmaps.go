package configmaps

import (
	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetConfigMapByName retrieves a ConfigMap by name in the specified namespace using the wrangler context for the given cluster
func GetConfigMapByName(client *rancher.Client, clusterID, namespace, configMapName string) (*corev1.ConfigMap, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	return clusterContext.Core.ConfigMap().Get(namespace, configMapName, metav1.GetOptions{})
}
