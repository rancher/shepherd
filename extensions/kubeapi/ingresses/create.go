package ingresses

import (
	"github.com/rancher/shepherd/clients/rancher"
	clusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateIngress is a helper function that uses the wrangler context to create an ingress in a specific namespace for a specific cluster.
func CreateIngress(client *rancher.Client, clusterID, ingressName, namespace string, ingressSpec *networkingv1.IngressSpec) (*networkingv1.Ingress, error) {
	wranglerCtx, err := clusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: namespace,
		},
		Spec: *ingressSpec,
	}

	createdIngress, err := wranglerCtx.Networking.Ingress().Create(ingress)
	if err != nil {
		return nil, err
	}

	return createdIngress, nil
}
