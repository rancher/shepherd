package statefulsets

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// UpdateStatefulSet updates an existing StatefulSet using wrangler context
func UpdateStatefulSet(client *rancher.Client, clusterID string, updatedSS *appsv1.StatefulSet, waitForReady bool) (*appsv1.StatefulSet, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	var updated *appsv1.StatefulSet
	var lastErr error

	err = kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		current, getErr := clusterContext.Apps.StatefulSet().Get(updatedSS.Namespace, updatedSS.Name, metav1.GetOptions{})
		if getErr != nil {
			lastErr = fmt.Errorf("failed to get StatefulSet %s/%s: %w", updatedSS.Namespace, updatedSS.Name, getErr)
			return false, nil
		}
		updatedSS.ResourceVersion = current.ResourceVersion
		updated, lastErr = clusterContext.Apps.StatefulSet().Update(updatedSS)
		if lastErr != nil {
			if errors.IsConflict(lastErr) {
				return false, nil
			}
			return false, lastErr
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("timed out updating StatefulSet %s/%s: %w", updatedSS.Namespace, updatedSS.Name, lastErr)
	}

	if waitForReady {
		err = WaitForStatefulSetReady(client, clusterID, updated.Namespace, updated.Name)
		if err != nil {
			return nil, err
		}
	}

	return updated, nil
}
