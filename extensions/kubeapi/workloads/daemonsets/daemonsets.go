package daemonsets

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// GetDaemonSetByName returns a DaemonSet by name and namespace using wrangler context.
func GetDaemonSetByName(client *rancher.Client, clusterID, daemonSetNamespace, daemonSetName string) (*appsv1.DaemonSet, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	ds, err := clusterContext.Apps.DaemonSet().Get(daemonSetNamespace, daemonSetName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get DaemonSet %s/%s: %w", daemonSetNamespace, daemonSetName, err)
	}

	return ds, nil
}

// WaitForDaemonSetReady waits until the DaemonSet is fully ready and rolled out.
func WaitForDaemonSetReady(client *rancher.Client, clusterID, daemonSetNamespace, daemonSetName string) error {
	return kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.FiveMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		ds, err := GetDaemonSetByName(client, clusterID, daemonSetNamespace, daemonSetName)
		if err != nil {
			return false, nil
		}

		return ds.Status.NumberReady == ds.Status.DesiredNumberScheduled &&
			ds.Status.UpdatedNumberScheduled == ds.Status.DesiredNumberScheduled, nil
	})
}
