package jobs

import (
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	batchv1 "k8s.io/api/batch/v1"
)

// CreateJobWithTemplate creates a Job in a cluster using the provided template and wrangler context.
func CreateJobWithTemplate(client *rancher.Client, clusterID string, jobTemplate *batchv1.Job, waitForActive bool) (*batchv1.Job, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	createdJob, err := clusterContext.Batch.Job().Create(jobTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to create Job: %w", err)
	}

	if waitForActive {
		err = WaitForJobActive(client, clusterID, createdJob.Namespace, createdJob.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to wait for job to become active: %w", err)
		}
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
			if err != nil {
				return err
			}

			return DeleteJob(adminClient, clusterID, createdJob.Namespace, createdJob.Name, true)
		})
	}

	return createdJob, nil
}
