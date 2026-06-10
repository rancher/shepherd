package pods

import (
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	corev1 "k8s.io/api/core/v1"
)

// CreatePodWithTemplate creates a pod in a cluster using the provided template and wrangler context. If waitForPod is true, it will wait for the pod to be running
func CreatePodWithTemplate(client *rancher.Client, clusterID string, podTemplate *corev1.Pod, waitForPod bool) (*corev1.Pod, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	createdPod, err := clusterContext.Core.Pod().Create(podTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %w", err)
	}

	if waitForPod {
		err = WaitForPodRunning(client, clusterID, createdPod.Namespace, createdPod.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to wait for pod to be running: %w", err)
		}
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
			if err != nil {
				return err
			}

			return DeletePod(adminClient, clusterID, createdPod.Namespace, createdPod.Name, true)
		})
	}

	return createdPod, nil
}
