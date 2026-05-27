package jobs

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
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

// WaitForJobComplete waits until the Job is complete using wrangler context and polling.
func WaitForJobComplete(client *rancher.Client, clusterID, jobNamespace, jobName string) error {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return err
	}

	return kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.FiveMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		job, err := clusterContext.Batch.Job().Get(jobNamespace, jobName, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}

		for _, c := range job.Status.Conditions {
			if c.Type == batchv1.JobComplete && c.Status == corev1.ConditionTrue {
				return true, nil
			}
		}
		return false, nil
	})
}
