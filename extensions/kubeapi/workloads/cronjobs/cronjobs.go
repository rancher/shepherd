package cronjobs

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// WaitForCronJobActive waits until the CronJob has at least one active job using wrangler context and polling.
func WaitForCronJobActive(client *rancher.Client, clusterID, namespaceName, cronJobName string) error {
	return kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.FiveMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		cronJob, err := GetCronJobByName(client, clusterID, namespaceName, cronJobName)
		if err != nil {
			return false, nil
		}

		if len(cronJob.Status.Active) > 0 {
			return true, nil
		}

		return false, nil
	})
}

// GetCronJobByName returns a CronJob by name and namespace using wrangler context.
func GetCronJobByName(client *rancher.Client, clusterID, cronJobNamespace, cronJobName string) (*batchv1.CronJob, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	cronJob, err := clusterContext.Batch.CronJob().Get(cronJobNamespace, cronJobName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get CronJob %s/%s: %w", cronJobNamespace, cronJobName, err)
	}

	return cronJob, nil
}
