package jobs

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

// GetJobByName returns a Job by name and namespace using wrangler context.
func GetJobByName(client *rancher.Client, clusterID, jobNamespace, jobName string) (*batchv1.Job, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	job, err := clusterContext.Batch.Job().Get(jobNamespace, jobName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get Job %s/%s: %w", jobNamespace, jobName, err)
	}

	return job, nil
}

// WaitForJobActive waits until the Job is active using wrangler context and polling.
func WaitForJobActive(client *rancher.Client, clusterID, jobNamespace, jobName string) error {
	return kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.FiveMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		job, err := GetJobByName(client, clusterID, jobNamespace, jobName)
		if err != nil {
			return false, nil
		}

		if job.Status.Active == 1 {
			return true, nil
		}

		return false, nil
	})
}
