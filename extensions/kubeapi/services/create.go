package services

import (
	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	corev1 "k8s.io/api/core/v1"
)

// CreateServiceWithTemplate creates a service using the provided template, respecting its name and metadata.
func CreateServiceWithTemplate(client *rancher.Client, clusterID string, serviceTemplate *corev1.Service) (*corev1.Service, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	service, err := clusterContext.Core.Service().Create(serviceTemplate)
	if err != nil {
		return nil, err
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
			if err != nil {
				return err
			}

			return DeleteService(adminClient, clusterID, service.Namespace, service.Name)
		})
	}

	return service, nil
}
