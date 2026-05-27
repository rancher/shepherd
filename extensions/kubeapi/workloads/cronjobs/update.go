package cronjobs

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

// UpdateCronJob updates an existing CronJob using wrangler context. If waitForActive is true, waits for the CronJob to become active after update.
func UpdateCronJob(client *rancher.Client, clusterID string, updatedCronJob *batchv1.CronJob, waitForActive bool) (*batchv1.CronJob, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	var updated *batchv1.CronJob
	var lastErr error

	err = kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		current, getErr := clusterContext.Batch.CronJob().Get(updatedCronJob.Namespace, updatedCronJob.Name, metav1.GetOptions{})
		if getErr != nil {
			lastErr = fmt.Errorf("failed to get CronJob %s/%s: %w", updatedCronJob.Namespace, updatedCronJob.Name, getErr)
			return false, nil
		}
		updatedCronJob.ResourceVersion = current.ResourceVersion
		updated, lastErr = clusterContext.Batch.CronJob().Update(updatedCronJob)
		if lastErr != nil {
			if errors.IsConflict(lastErr) {
				return false, nil
			}
			return false, lastErr
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("timed out updating CronJob %s/%s: %w", updatedCronJob.Namespace, updatedCronJob.Name, lastErr)
	}

	if waitForActive {
		err = WaitForCronJobActive(client, clusterID, updatedCronJob.Namespace, updatedCronJob.Name)
		if err != nil {
			return updated, fmt.Errorf("updated but did not become active: %w", err)
		}
	}

	return updated, nil
}
