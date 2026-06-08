package pods

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

// DeletePod deletes the specified pod from the given namespace using wrangler context. If waitForDeletion is true, it will wait for the pod to be deleted
func DeletePod(client *rancher.Client, clusterID, namespace, podName string, waitForDeletion bool) error {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return err
	}

	err = clusterContext.Core.Pod().Delete(namespace, podName, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if waitForDeletion {
		err = WaitForPodDeleted(client, clusterID, namespace, podName)
		if err != nil {
			return fmt.Errorf("error waiting for pod deletion: %w", err)
		}
	}

	return nil
}

// WaitForPodDeleted waits until the specified pod is deleted from the cluster
func WaitForPodDeleted(client *rancher.Client, clusterID, namespace, podName string) error {
	return kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(context.Context) (bool, error) {
		_, err := GetPodByName(client, clusterID, namespace, podName)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	})
}
