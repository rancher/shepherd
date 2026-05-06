package statefulsets

import (
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	appsv1 "k8s.io/api/apps/v1"
)

// CreateStatefulSetWithTemplate creates a StatefulSet in a cluster using the provided template and wrangler context.
func CreateStatefulSetWithTemplate(client *rancher.Client, clusterID string, ssTemplate *appsv1.StatefulSet, waitForReady bool) (*appsv1.StatefulSet, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	createdSS, err := clusterContext.Apps.StatefulSet().Create(ssTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to create StatefulSet: %w", err)
	}

	if waitForReady {
		err = WaitForStatefulSetReady(client, clusterID, createdSS.Namespace, createdSS.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to wait for StatefulSet to be ready: %w", err)
		}
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			return DeleteStatefulSet(client, clusterID, createdSS.Namespace, createdSS.Name, true)
		})
	}

	return createdSS, nil
}
