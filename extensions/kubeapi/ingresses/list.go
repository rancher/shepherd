package ingresses

import (
	"github.com/rancher/shepherd/clients/rancher"
	clusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressList is a struct that contains a list of Ingresses.
type IngressList struct {
	Items []networkingv1.Ingress
}

// ListIngresses is a helper function that uses the wrangler context to list ingresses in a namespace for a specific cluster and returns an IngressList.
func ListIngresses(client *rancher.Client, clusterID, namespace string, listOpts metav1.ListOptions) (*IngressList, error) {
	clusterContext, err := clusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	ingressList, err := clusterContext.Networking.Ingress().List(namespace, listOpts)
	if err != nil {
		return nil, err
	}

	return &IngressList{Items: ingressList.Items}, nil
}

// Names returns each Ingress name in the list as a new slice of strings.
func (list *IngressList) Names() []string {
	var names []string
	for _, ing := range list.Items {
		names = append(names, ing.Name)
	}
	return names
}
