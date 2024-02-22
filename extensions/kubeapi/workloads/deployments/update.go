package deployments

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/unstructured"
	"github.com/rancher/shepherd/pkg/api/scheme"
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UpdateDeployment is a helper function that uses the dynamic client to update a deployment in a namespace for a specific cluster.
func UpdateDeployment(client *rancher.Client, clusterID string, namespaceName string, updatedDeployment *appv1.Deployment) (*appv1.Deployment, error) {
	dynamicClient, err := client.GetDownStreamClusterClient(clusterID)
	if err != nil {
		return nil, err
	}

	deploymentResource := dynamicClient.Resource(DeploymentGroupVersionResource).Namespace(namespaceName)

	deploymentUnstructured, err := deploymentResource.Get(context.TODO(), updatedDeployment.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	currentDeployment := &appv1.Deployment{}
	err = scheme.Scheme.Convert(deploymentUnstructured, currentDeployment, deploymentUnstructured.GroupVersionKind())
	if err != nil {
		return nil, err
	}

	updatedDeployment.ObjectMeta.ResourceVersion = currentDeployment.ObjectMeta.ResourceVersion

	unstructuredResp, err := deploymentResource.Update(context.TODO(), unstructured.MustToUnstructured(updatedDeployment), metav1.UpdateOptions{})
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
