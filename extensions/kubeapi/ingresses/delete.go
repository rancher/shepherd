package ingresses

import (
	"github.com/rancher/shepherd/clients/rancher"
	clusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeleteIngress is a helper function that uses the wrangler context to delete an ingress in a specific namespace for a specific cluster.
func DeleteIngress(client *rancher.Client, clusterID, namespace, ingressName string) error {
	wranglerCtx, err := clusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return err
	}

	err = wranglerCtx.Networking.Ingress().Delete(namespace, ingressName, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}
