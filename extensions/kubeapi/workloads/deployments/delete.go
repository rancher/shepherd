package deployments

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// DeleteDeployment is a helper function that uses the dynamic client to delete a deployment in a namespace for a specific cluster.
func DeleteDeployment(client *rancher.Client, clusterID string, namespaceName string, deploymentName string) error {
	dynamicClient, err := client.GetDownStreamClusterClient(clusterID)
	if err != nil {
		return err
	}

	deploymentResource := dynamicClient.Resource(DeploymentGroupVersionResource).Namespace(namespaceName)

	err = deploymentResource.Delete(context.TODO(), deploymentName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaults.OneMinuteTimeout)
	defer cancel()

	err = kwait.PollUntilContextCancel(ctx, defaults.FiveHundredMillisecondTimeout, false, func(ctx context.Context) (done bool, err error) {
		deploymentList, err := ListDeployments(client, clusterID, namespaceName, metav1.ListOptions{
			FieldSelector: "metadata.name=" + deploymentName,
		})
		if err != nil {
			return false, err
		}

		if len(deploymentList.Items) == 0 {
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return err
	}

	return nil
}
