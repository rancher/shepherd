package ingresses

import (
	"github.com/rancher/shepherd/clients/rancher"
	clusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListIngresses is a helper function that uses the wrangler context to list ingresses in a namespace for a specific cluster.
func ListIngresses(client *rancher.Client, clusterID, namespace string, listOpts metav1.ListOptions) (*networkingv1.IngressList, error) {
	wranglerCtx, err := clusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	ingressList, err := wranglerCtx.Networking.Ingress().List(namespace, listOpts)
	if err != nil {
		return nil, err
	}

	return ingressList, nil
}
