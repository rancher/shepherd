package jobs

import (
	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// JobList is a struct that contains a list of Jobs.
type JobList struct {
	Items []batchv1.Job
}

// ListJobs returns Jobs in a specific namespace using wrangler context and returns a JobList.
func ListJobs(client *rancher.Client, clusterID, namespace string, listOpts metav1.ListOptions) (*JobList, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	jobList, err := clusterContext.Batch.Job().List(namespace, listOpts)
	if err != nil {
		return nil, err
	}

	return &JobList{Items: jobList.Items}, nil
}

// Names returns each Job name in the list as a new slice of strings.
func (list *JobList) Names() []string {
	var names []string
	for _, j := range list.Items {
		names = append(names, j.Name)
	}
	return names
}
