package deployments

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// GetDeploymentByName returns a Deployment by name and namespace using wrangler context.
func GetDeploymentByName(client *rancher.Client, clusterID, deploymentNamespace, deploymentName string) (*appsv1.Deployment, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	deployment, err := clusterContext.Apps.Deployment().Get(deploymentNamespace, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get Deployment %s/%s: %w", deploymentNamespace, deploymentName, err)
	}

	return deployment, nil
}

// WaitForDeploymentActive waits until the Deployment has all available replicas using wrangler context and polling.
func WaitForDeploymentActive(client *rancher.Client, clusterID, deploymentNamespace, deploymentName string) error {
	return kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.FiveMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		deployment, err := GetDeploymentByName(client, clusterID, deploymentNamespace, deploymentName)
		if err != nil {
			return false, nil
		}

		if deployment.Spec.Replicas == nil {
			return false, nil
		}

		desired := *deployment.Spec.Replicas
		return deployment.Status.UpdatedReplicas == desired &&
			deployment.Status.AvailableReplicas == desired &&
			deployment.Status.Replicas == desired, nil
	})
}
