package statefulsets

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

// DeleteStatefulSet deletes a StatefulSet in a cluster using wrangler context and waits for deletion.
func DeleteStatefulSet(client *rancher.Client, clusterID, namespace, name string, waitForDelete bool) error {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return err
	}

	err = clusterContext.Apps.StatefulSet().Delete(namespace, name, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if waitForDelete {
		err = WaitForStatefulSetDeleted(client, clusterID, namespace, name)
		if err != nil {
			return fmt.Errorf("timed out waiting for statefulset %s to be deleted: %w", name, err)
		}
	}

	return nil
}

// WaitForStatefulSetDeleted waits until the specified StatefulSet is deleted from the cluster.
func WaitForStatefulSetDeleted(client *rancher.Client, clusterID, namespace, name string) error {
	return kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		_, err := GetStatefulSetByName(client, clusterID, namespace, name)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	})
}
