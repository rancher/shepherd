package daemonsets

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

// UpdateDaemonSet updates an existing DaemonSet using wrangler context. If waitForReady is true, it will wait for the DaemonSet to be ready after the update.
func UpdateDaemonSet(client *rancher.Client, clusterID string, updatedDS *appsv1.DaemonSet, waitForReady bool) (*appsv1.DaemonSet, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	var updated *appsv1.DaemonSet
	var lastErr error

	err = kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		current, getErr := GetDaemonSetByName(client, clusterID, updatedDS.Namespace, updatedDS.Name)
		if getErr != nil {
			lastErr = fmt.Errorf("failed to get DaemonSet %s/%s: %w", updatedDS.Namespace, updatedDS.Name, getErr)
			return false, nil
		}
		updatedDS.ResourceVersion = current.ResourceVersion
		updated, lastErr = clusterContext.Apps.DaemonSet().Update(updatedDS)
		if lastErr != nil {
			if errors.IsConflict(lastErr) {
				return false, nil
			}
			return false, lastErr
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("timed out updating DaemonSet %s/%s: %w", updatedDS.Namespace, updatedDS.Name, lastErr)
	}

	if waitForReady {
		err = WaitForDaemonSetReady(client, clusterID, updated.Namespace, updated.Name)
		if err != nil {
			return nil, err
		}
	}

	return updated, nil
}
