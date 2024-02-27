package deployments

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/api/scheme"
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetDeployment is a helper function that uses the dynamic client to get a deployment in a cluster with its get options.
func GetDeployment(client *rancher.Client, clusterID, name, namespace string, getOpts metav1.GetOptions) (*appv1.Deployment, error) {

	dynamicClient, err := client.GetDownStreamClusterClient(clusterID)
	if err != nil {
		return nil, err
	}

	deploymentResource := dynamicClient.Resource(DeploymentGroupVersionResource).Namespace(namespace)
	unstructuredResp, err := deploymentResource.Get(context.TODO(), name, getOpts)
	if err != nil {
		return nil, err
	}

	newDeployment := &appv1.Deployment{}
	err = scheme.Scheme.Convert(unstructuredResp, newDeployment, unstructuredResp.GroupVersionKind())
	if err != nil {
		return nil, err
	}

	return newDeployment, nil
}
