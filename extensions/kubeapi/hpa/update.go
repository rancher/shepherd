package hpa

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// UpdateHPA updates an existing HorizontalPodAutoscaler using wrangler context. If waitForActive is true, it will wait for the HPA to be active after the update.
func UpdateHPA(client *rancher.Client, clusterID string, updatedHPA *autoscalingv2.HorizontalPodAutoscaler, waitForActive bool) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	var updated *autoscalingv2.HorizontalPodAutoscaler
	var lastErr error

	err = kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		current, getErr := clusterContext.Autoscaling.HorizontalPodAutoscaler().Get(updatedHPA.Namespace, updatedHPA.Name, metav1.GetOptions{})
		if getErr != nil {
			lastErr = fmt.Errorf("failed to get HPA %s/%s: %w", updatedHPA.Namespace, updatedHPA.Name, getErr)
			return false, nil
		}
		updatedHPA.ResourceVersion = current.ResourceVersion
		updated, lastErr = clusterContext.Autoscaling.HorizontalPodAutoscaler().Update(updatedHPA)
		if lastErr != nil {
			if errors.IsConflict(lastErr) {
				return false, nil
			}
			return false, lastErr
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("timed out updating HPA %s/%s: %w", updatedHPA.Namespace, updatedHPA.Name, lastErr)
	}

	if waitForActive {
		err = WaitForHPAActive(client, clusterID, updated.Namespace, updated.Name, updated.Status.DesiredReplicas)
		if err != nil {
			return nil, err
		}
	}

	return updated, nil
}
