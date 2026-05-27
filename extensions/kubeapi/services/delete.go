package services

import (
	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeleteService deletes a service in the specified cluster and namespace using the wrangler context
func DeleteService(client *rancher.Client, clusterID, namespace, serviceName string) error {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return err
	}

	err = clusterContext.Core.Service().Delete(namespace, serviceName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}
