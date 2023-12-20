package projects

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeleteProject is a helper function that uses the dynamic client to delete a Project from a cluster.
func DeleteProject(client *rancher.Client, projectNamespace string, projectName string) error {
	dynamicClient, err := client.GetDownStreamClusterClient(LocalCluster)
	if err != nil {
		return err
	}

	projectResource := dynamicClient.Resource(ProjectGroupVersionResource).Namespace(projectNamespace)

	err = projectResource.Delete(context.TODO(), projectName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}
