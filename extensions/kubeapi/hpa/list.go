package hpa

import (
	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// List is a struct that contains a list of horizontal pod autoscalers.
type List struct {
	Items []autoscalingv2.HorizontalPodAutoscaler
}

// ListHPAs is a helper function that uses the wrangler client to list horizontal pod autoscalers on a namespace for a specific cluster with its list options.
func ListHPAs(client *rancher.Client, clusterID, namespace string, listOpts metav1.ListOptions) (*List, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	hpas, err := clusterContext.Autoscaling.HorizontalPodAutoscaler().List(namespace, listOpts)
	if err != nil {
		return nil, err
	}

	return &List{Items: hpas.Items}, nil
}

// Names is a method that accepts List as a receiver, returns each HPA name in the list as a new slice of strings.
func (list *List) Names() []string {
	var hpaNames []string
	for _, hpa := range list.Items {
		hpaNames = append(hpaNames, hpa.Name)
	}
	return hpaNames
}
