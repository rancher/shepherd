package ingresses

import (
	"github.com/rancher/shepherd/clients/rancher"
	clusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UpdateIngress is a helper function that uses the wrangler context to update an existing ingress with new specifications.
func UpdateIngress(client *rancher.Client, clusterID, namespace string, existingIngress, updatedIngress *networkingv1.Ingress) (*networkingv1.Ingress, error) {
	wranglerCtx, err := clusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	currentIngress, err := wranglerCtx.Networking.Ingress().Get(namespace, existingIngress.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if updatedIngress.Name == "" {
		updatedIngress.Name = existingIngress.Name
	}
	if updatedIngress.Namespace == "" {
		updatedIngress.Namespace = namespace
	}

	updatedIngress.ResourceVersion = currentIngress.ResourceVersion

	newIngress, err := wranglerCtx.Networking.Ingress().Update(updatedIngress)
	if err != nil {
		return nil, err
	}

	return newIngress, nil
}
