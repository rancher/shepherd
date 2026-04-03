package deployments

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/extensions/kubeapi/cluster"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

const (
	RancherDeploymentName      = "rancher"
	RancherDeploymentNamespace = "cattle-system"
)

// WaitForDeploymentActive polls until the deployment has all replicas updated, ready, and available.
func WaitForDeploymentActive(client *rancher.Client, clusterID, namespaceName, deploymentName string) error {
	wranglerContext, err := cluster.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return fmt.Errorf("getting wrangler context for cluster %s: %w", clusterID, err)
	}
	return kwait.PollUntilContextTimeout(
		context.Background(), defaults.TenSecondTimeout, defaults.FiveMinuteTimeout, false,
		func(ctx context.Context) (bool, error) {
			d, err := wranglerContext.Apps.Deployment().Get(namespaceName, deploymentName, metav1.GetOptions{})
			if err != nil {
				logrus.Debugf("waiting for deployment %s/%s: %v", namespaceName, deploymentName, err)
				return false, nil
			}
			if d.Spec.Replicas == nil {
				return false, nil
			}
			desired := *d.Spec.Replicas
			return d.Status.UpdatedReplicas == desired &&
				d.Status.ReadyReplicas == desired &&
				d.Status.AvailableReplicas == desired &&
				d.Status.Replicas == desired, nil
		},
	)
}
