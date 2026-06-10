package cronjobs

import (
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	batchv1 "k8s.io/api/batch/v1"
)

// CreateCronJobWithTemplate creates a CronJob in a cluster using the provided template and wrangler context. If waitForActive is true, it will wait for the CronJob to have an active job
func CreateCronJobWithTemplate(client *rancher.Client, clusterID string, cronJobTemplate *batchv1.CronJob, waitForActive bool) (*batchv1.CronJob, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	createdCronJob, err := clusterContext.Batch.CronJob().Create(cronJobTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to create CronJob: %w", err)
	}

	if waitForActive {
		err = WaitForCronJobActive(client, clusterID, createdCronJob.Namespace, createdCronJob.Name)
		if err != nil {
			return nil, err
		}
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
			if err != nil {
				return err
			}

			return DeleteCronJob(adminClient, clusterID, createdCronJob.Namespace, createdCronJob.Name, true)
		})
	}

	return createdCronJob, nil
}
