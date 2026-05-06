package configmaps

import (
	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigMapList is a struct that contains a list of ConfigMaps.
type ConfigMapList struct {
	Items []corev1.ConfigMap
}

// ListConfigMaps returns ConfigMaps in a specific namespace using wrangler context and returns a ConfigMapList.
func ListConfigMaps(client *rancher.Client, clusterID, namespace string, listOpts metav1.ListOptions) (*ConfigMapList, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	configMapList, err := clusterContext.Core.ConfigMap().List(namespace, listOpts)
	if err != nil {
		return nil, err
	}

	return &ConfigMapList{Items: configMapList.Items}, nil
}

// Names returns each ConfigMap name in the list as a new slice of strings.
func (list *ConfigMapList) Names() []string {
	var names []string
	for _, cm := range list.Items {
		names = append(names, cm.Name)
	}
	return names
}
