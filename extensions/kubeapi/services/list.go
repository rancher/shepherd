package services

import (
	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceList is a struct that contains a list of Services.
type ServiceList struct {
	Items []corev1.Service
}

// ListServices is a helper function that uses the wrangler context to list services in a namespace for a specific cluster and returns a ServiceList.
func ListServices(client *rancher.Client, clusterID, namespace string, listOpts metav1.ListOptions) (*ServiceList, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	services, err := clusterContext.Core.Service().List(namespace, listOpts)
	if err != nil {
		return nil, err
	}

	return &ServiceList{Items: services.Items}, nil
}

// Names returns each Service name in the list as a new slice of strings.
func (list *ServiceList) Names() []string {
	var names []string
	for _, s := range list.Items {
		names = append(names, s.Name)
	}
	return names
}
