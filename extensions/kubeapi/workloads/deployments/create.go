package deployments

import (
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	appsv1 "k8s.io/api/apps/v1"
)

// CreateDeploymentWithTemplate creates a Deployment in a cluster using the provided template and wrangler context.
func CreateDeploymentWithTemplate(client *rancher.Client, clusterID string, deploymentTemplate *appsv1.Deployment, waitForActive bool) (*appsv1.Deployment, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	createdDeployment, err := clusterContext.Apps.Deployment().Create(deploymentTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to create Deployment: %w", err)
	}

	if waitForActive {
		err = WaitForDeploymentActive(client, clusterID, createdDeployment.Namespace, createdDeployment.Name)
		if err != nil {
			return nil, fmt.Errorf("error waiting for Deployment to become active: %w", err)
		}
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			return DeleteDeployment(client, clusterID, createdDeployment.Namespace, createdDeployment.Name, true)
		})
	}

	return createdDeployment, nil
}
