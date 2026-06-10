package jobs

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// DeleteJob deletes a Job in a cluster using wrangler context and waits for deletion.
func DeleteJob(client *rancher.Client, clusterID, jobNamespace, jobName string, waitForDeletion bool) error {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return err
	}

	propagationPolicy := metav1.DeletePropagationBackground
	err = clusterContext.Batch.Job().Delete(jobNamespace, jobName, &metav1.DeleteOptions{PropagationPolicy: &propagationPolicy})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if waitForDeletion {
		return WaitForJobDeleted(client, clusterID, jobNamespace, jobName)
	}

	return nil
}

// WaitForJobDeleted waits until the specified Job is deleted from the cluster.
func WaitForJobDeleted(client *rancher.Client, clusterID, jobNamespace, jobName string) error {
	return kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		_, err := GetJobByName(client, clusterID, jobNamespace, jobName)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	})
}
