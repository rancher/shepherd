package hpa

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// GetHPAByName retrieves a HorizontalPodAutoscaler by name in the given namespace using wrangler context.
func GetHPAByName(client *rancher.Client, clusterID, hpaNamespace, hpaName string) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	hpa, err := clusterContext.Autoscaling.HorizontalPodAutoscaler().Get(hpaNamespace, hpaName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get HPA: %w", err)
	}
	return hpa, nil
}

// WaitForHPAActive waits until the HPA has a current number of replicas matching the desired state.
func WaitForHPAActive(client *rancher.Client, clusterID, hpaNamespace, hpaName string, desiredReplicas int32) error {
	return kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.FiveMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		hpa, err := GetHPAByName(client, clusterID, hpaNamespace, hpaName)
		if err != nil {
			return false, nil
		}
		for _, condition := range hpa.Status.Conditions {
			if condition.Type == autoscalingv2.ScalingActive && condition.Status == "True" {
				return true, nil
			}
		}
		return false, nil
	})
}
