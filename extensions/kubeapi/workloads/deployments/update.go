package deployments

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// UpdateDeployment updates an existing Deployment using wrangler context. If waitForActive is true, it will wait for the Deployment to be active after the update.
func UpdateDeployment(client *rancher.Client, clusterID string, updatedDeployment *appsv1.Deployment, waitForActive bool) (*appsv1.Deployment, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	var updated *appsv1.Deployment
	var lastErr error

	err = kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		current, getErr := GetDeploymentByName(client, clusterID, updatedDeployment.Namespace, updatedDeployment.Name)
		if getErr != nil {
			lastErr = fmt.Errorf("failed to get Deployment %s/%s: %w", updatedDeployment.Namespace, updatedDeployment.Name, getErr)
			return false, nil
		}
		updatedDeployment.ResourceVersion = current.ResourceVersion
		updated, lastErr = clusterContext.Apps.Deployment().Update(updatedDeployment)
		if lastErr != nil {
			if errors.IsConflict(lastErr) {
				return false, nil
			}
			return false, lastErr
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("timed out updating Deployment %s/%s: %w", updatedDeployment.Namespace, updatedDeployment.Name, lastErr)
	}

	if waitForActive {
		err = WaitForDeploymentActive(client, clusterID, updatedDeployment.Namespace, updatedDeployment.Name)
		if err != nil {
			return nil, fmt.Errorf("error waiting for Deployment %s/%s to become active: %w", updatedDeployment.Namespace, updatedDeployment.Name, err)
		}
	}

	return updated, nil
}
