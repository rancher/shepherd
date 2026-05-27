package statefulsets

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

// GetStatefulSetByName returns a StatefulSet by name and namespace using wrangler context.
func GetStatefulSetByName(client *rancher.Client, clusterID, statefulSetNamespace, statefulSetName string) (*appsv1.StatefulSet, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	ss, err := clusterContext.Apps.StatefulSet().Get(statefulSetNamespace, statefulSetName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get StatefulSet %s/%s: %w", statefulSetNamespace, statefulSetName, err)
	}

	return ss, nil
}

// WaitForStatefulSetReady waits until the StatefulSet has all ready replicas using wrangler context and polling.
func WaitForStatefulSetReady(client *rancher.Client, clusterID, statefulSetNamespace, statefulSetName string) error {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return err
	}

	return kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.FiveMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		ss, err := clusterContext.Apps.StatefulSet().Get(statefulSetNamespace, statefulSetName, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}

		if ss.Spec.Replicas != nil && ss.Status.ReadyReplicas == *ss.Spec.Replicas {
			return true, nil
		}
		return false, nil
	})
}
