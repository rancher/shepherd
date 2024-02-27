package deployments

import (
	"context"
	"fmt"
	"time"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/api/scheme"
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func PatchDeployment(client *rancher.Client, clusterID, deploymentName, namespace string, data string, patchType types.PatchType) (*appv1.Deployment, error) {
	dynamicClient, err := client.GetDownStreamClusterClient(clusterID)
	if err != nil {
		return nil, err
	}
	deploymentResource := dynamicClient.Resource(DeploymentGroupVersionResource).Namespace(namespace)

	unstructuredResp, err := deploymentResource.Patch(context.TODO(), deploymentName, patchType, []byte(data), metav1.PatchOptions{})
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

// PatchDeploymentFromYAML is a helper function that uses the dynamic client to patch a deployment in a namespace for a specific cluster.
// Different merge strategies are supported based on the PatchType.
func PatchDeploymentFromYAML(client *rancher.Client, clusterID, deploymentName, namespace string, rawYAML []byte, patchType types.PatchType) (*appv1.Deployment, error) {
	dynamicClient, err := client.GetDownStreamClusterClient(clusterID)
	if err != nil {
		return nil, err
	}
	deploymentResource := dynamicClient.Resource(DeploymentGroupVersionResource).Namespace(namespace)

	rawJSON, err := yaml.ToJSON(rawYAML)
	if err != nil {
		return nil, err
	}

	unstructuredResp, err := deploymentResource.Patch(context.TODO(), deploymentName, patchType, rawJSON, metav1.PatchOptions{})
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

func RestartDeployment(client *rancher.Client, clusterID, deploymentName, namespace string) (*appv1.Deployment, error) {
	data := fmt.Sprintf(`{"spec": {"template": {"metadata": {"annotations": {"kubectl.kubernetes.io/restartedAt": "%s"}}}}}`, time.Now().Format("20060102150405"))
	return PatchDeploymentFromYAML(client, clusterID, deploymentName, namespace, []byte(data), types.StrategicMergePatchType)
}
