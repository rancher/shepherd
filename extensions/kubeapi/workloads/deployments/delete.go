package deployments

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// DeleteDeployment deletes a Deployment in a cluster using wrangler context and waits for deletion. If waitForDeletion is true, it will wait for the Deployment to be deleted
func DeleteDeployment(client *rancher.Client, clusterID, namespace, deploymentName string, waitForDeletion bool) error {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return err
	}

	err = clusterContext.Apps.Deployment().Delete(namespace, deploymentName, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if waitForDeletion {
		return WaitForDeploymentDeleted(client, clusterID, namespace, deploymentName)
	}
	return nil
}

// WaitForDeploymentDeleted waits until the specified Deployment is deleted from the cluster.
func WaitForDeploymentDeleted(client *rancher.Client, clusterID, deploymentNamespace, deploymentName string) error {
	return kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		_, err := GetDeploymentByName(client, clusterID, deploymentNamespace, deploymentName)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	})
}
