package deployments

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	clusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

const (
	RancherDeploymentNamespace = "cattle-system"
	RancherDeploymentName      = "rancher"
)

// WaitForDeploymentActive uses wrangler context to wait for a deployment to become active
func WaitForDeploymentActive(client *rancher.Client, clusterID, namespaceName, deploymentName string) error {
	wranglerContext, err := clusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return err
	}

	return kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.FiveMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		deployment, err := wranglerContext.Apps.Deployment().Get(namespaceName, deploymentName, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}

		desired := *deployment.Spec.Replicas

		if deployment.Status.ReadyReplicas != desired {
			return false, nil
		}

		return true, nil
	},
	)
}
