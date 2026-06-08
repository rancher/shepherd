package daemonsets

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// DeleteDaemonSet deletes a DaemonSet in a cluster using wrangler context and waits for deletion. If waitForDelete is true, it will wait for the DaemonSet to be deleted
func DeleteDaemonSet(client *rancher.Client, clusterID, daemonSetNamespace, daemonSetName string, waitForDelete bool) error {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return fmt.Errorf("failed to get cluster context: %w", err)
	}

	err = clusterContext.Apps.DaemonSet().Delete(daemonSetNamespace, daemonSetName, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if waitForDelete {
		err = WaitForDaemonSetDeletion(client, clusterID, daemonSetNamespace, daemonSetName)
		if err != nil {
			return fmt.Errorf("failed to wait for daemonset to delete: %w", err)
		}
	}

	return nil
}

// WaitForDaemonSetDeletion waits until the specified DaemonSet is deleted from the cluster.
func WaitForDaemonSetDeletion(client *rancher.Client, clusterID, daemonSetNamespace, daemonSetName string) error {
	return kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(context.Context) (bool, error) {
		_, err := GetDaemonSetByName(client, clusterID, daemonSetNamespace, daemonSetName)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	})
}
