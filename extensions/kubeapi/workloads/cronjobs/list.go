package cronjobs

import (
	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CronJobList is a struct that contains a list of CronJobs.
type CronJobList struct {
	Items []batchv1.CronJob
}

// ListCronJobs is a helper to list cronjobs in a namespace using wrangler context and returns a CronJobList.
func ListCronJobs(client *rancher.Client, clusterID, namespace string, listOpts metav1.ListOptions) (*CronJobList, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	cronJobList, err := clusterContext.Batch.CronJob().List(namespace, listOpts)
	if err != nil {
		return nil, err
	}

	return &CronJobList{Items: cronJobList.Items}, nil
}

// Names returns each CronJob name in the list as a new slice of strings.
func (list *CronJobList) Names() []string {
	var names []string
	for _, cj := range list.Items {
		names = append(names, cj.Name)
	}
	return names
}
