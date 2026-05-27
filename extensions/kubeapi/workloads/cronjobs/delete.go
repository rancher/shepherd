package cronjobs

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// DeleteCronJob deletes a CronJob in a cluster using wrangler context.
func DeleteCronJob(client *rancher.Client, clusterID, cronJobnamespace, cronJobName string, waitForDelete bool) error {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return fmt.Errorf("failed to get cluster context: %w", err)
	}

	err = clusterContext.Batch.CronJob().Delete(cronJobnamespace, cronJobName, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete CronJob: %w", err)
	}

	if waitForDelete {
		err = WaitForCronJobDeletion(client, clusterID, cronJobnamespace, cronJobName)
		if err != nil {
			return fmt.Errorf("failed to wait for cronjob to delete: %w", err)
		}
	}

	return nil
}

// WaitForCronJobDeletion is a helper to wait for cronjob to delete
func WaitForCronJobDeletion(client *rancher.Client, clusterID, cronJobNamespace, cronJobName string) error {
	err := kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, pollErr error) {
		_, pollErr = GetCronJobByName(client, clusterID, cronJobNamespace, cronJobName)
		if pollErr != nil {
			if k8serrors.IsNotFound(pollErr) {
				return true, nil
			}
			return false, pollErr
		}
		return false, nil
	})

	if err != nil {
		return fmt.Errorf("failed to wait for cronjob to delete: %w", err)
	}

	return nil
}
