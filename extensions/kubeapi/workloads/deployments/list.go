package deployments

import (
	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeploymentList is a struct that contains a list of Deployments.
type DeploymentList struct {
	Items []appsv1.Deployment
}

// ListDeployments returns Deployments in a specific namespace using wrangler context and returns a DeploymentList.
func ListDeployments(client *rancher.Client, clusterID, namespace string, listOpts metav1.ListOptions) (*DeploymentList, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	deploymentList, err := clusterContext.Apps.Deployment().List(namespace, listOpts)
	if err != nil {
		return nil, err
	}

	return &DeploymentList{Items: deploymentList.Items}, nil
}

// Names returns each Deployment name in the list as a new slice of strings.
func (list *DeploymentList) Names() []string {
	var names []string
	for _, d := range list.Items {
		names = append(names, d.Name)
	}
	return names
}
