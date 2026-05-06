package pods

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// UpdatePod updates an existing Pod using wrangler context. If waitForPod is true, it will wait for the pod to be running after the update
func UpdatePod(client *rancher.Client, clusterID string, updatedPod *corev1.Pod, waitForPod bool) (*corev1.Pod, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	var updated *corev1.Pod
	var lastErr error

	err = kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		current, getErr := clusterContext.Core.Pod().Get(updatedPod.Namespace, updatedPod.Name, metav1.GetOptions{})
		if getErr != nil {
			lastErr = fmt.Errorf("failed to get Pod %s/%s: %w", updatedPod.Namespace, updatedPod.Name, getErr)
			return false, nil
		}
		updatedPod.ResourceVersion = current.ResourceVersion
		updated, lastErr = clusterContext.Core.Pod().Update(updatedPod)
		if lastErr != nil {
			if errors.IsConflict(lastErr) {
				return false, nil
			}
			return false, lastErr
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("timed out updating Pod %s/%s: %w", updatedPod.Namespace, updatedPod.Name, lastErr)
	}

	if waitForPod {
		err = WaitForPodRunning(client, clusterID, updated.Namespace, updated.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to wait for pod to be running: %w", err)
		}
	}

	return updated, nil
}
