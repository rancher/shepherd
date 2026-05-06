package daemonsets

import (
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	appsv1 "k8s.io/api/apps/v1"
)

// CreateDaemonSetWithTemplate creates a DaemonSet in a cluster using the provided template and wrangler context. If waitForReady is true, it will wait for the DaemonSet to be ready
func CreateDaemonSetWithTemplate(client *rancher.Client, clusterID string, daemonSetTemplate *appsv1.DaemonSet, waitForReady bool) (*appsv1.DaemonSet, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	createdDaemonset, err := clusterContext.Apps.DaemonSet().Create(daemonSetTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to create DaemonSet: %w", err)
	}

	if waitForReady {
		err = WaitForDaemonSetReady(client, clusterID, createdDaemonset.Namespace, createdDaemonset.Name)
		if err != nil {
			return nil, err
		}
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			return DeleteDaemonSet(client, clusterID, createdDaemonset.Namespace, createdDaemonset.Name, true)
		})
	}

	return createdDaemonset, nil
}
