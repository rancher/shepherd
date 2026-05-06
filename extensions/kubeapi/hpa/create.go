package hpa

import (
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
)

// CreateHPA creates a HorizontalPodAutoscaler in the given namespace using wrangler context. If waitForActive is true, it will wait for the HPA to be active after creation.
func CreateHPA(client *rancher.Client, clusterID string, hpa *autoscalingv2.HorizontalPodAutoscaler, waitForActive bool) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	createdHPA, err := clusterContext.Autoscaling.HorizontalPodAutoscaler().Create(hpa)
	if err != nil {
		return nil, fmt.Errorf("failed to create HPA: %w", err)
	}

	if waitForActive {
		err = WaitForHPAActive(client, clusterID, createdHPA.Namespace, createdHPA.Name, createdHPA.Status.DesiredReplicas)
		if err != nil {
			return createdHPA, fmt.Errorf("timed out waiting for HPA to be ready: %w", err)
		}
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			return DeleteHPA(client, clusterID, createdHPA.Namespace, createdHPA.Name, true)
		})
	}

	return createdHPA, nil
}
