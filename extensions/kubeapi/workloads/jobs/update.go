package jobs

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// UpdateJob updates an existing Job using wrangler context
func UpdateJob(client *rancher.Client, clusterID string, updatedJob *batchv1.Job, waitForCompletion bool) (*batchv1.Job, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	var updated *batchv1.Job
	var lastErr error

	err = kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		current, getErr := clusterContext.Batch.Job().Get(updatedJob.Namespace, updatedJob.Name, metav1.GetOptions{})
		if getErr != nil {
			lastErr = fmt.Errorf("failed to get Job %s/%s: %w", updatedJob.Namespace, updatedJob.Name, getErr)
			return false, nil
		}
		updatedJob.ResourceVersion = current.ResourceVersion
		updated, lastErr = clusterContext.Batch.Job().Update(updatedJob)
		if lastErr != nil {
			if errors.IsConflict(lastErr) {
				return false, nil
			}
			return false, lastErr
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("timed out updating Job %s/%s: %w", updatedJob.Namespace, updatedJob.Name, lastErr)
	}

	if waitForCompletion {
		err = WaitForJobComplete(client, clusterID, updated.Namespace, updated.Name)
		if err != nil {
			return nil, err
		}
	}

	return updated, nil
}
