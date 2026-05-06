package services

import (
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetServiceByName retrieves a service by name from a specific namespace in the given cluster using the wrangler context
func GetServiceByName(client *rancher.Client, clusterID, namespaceName, serviceName string) (*corev1.Service, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	service, err := clusterContext.Core.Service().Get(namespaceName, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service %s/%s: %w", namespaceName, serviceName, err)
	}

	return service, nil
}
