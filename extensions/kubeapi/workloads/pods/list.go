package pods

import (
	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodList is a struct that contains a list of Pods.
type PodList struct {
	Items []corev1.Pod
}

// ListPods returns Pods in a specific namespace using wrangler context and returns a PodList.
func ListPods(client *rancher.Client, clusterID, namespace string, listOpts metav1.ListOptions) (*PodList, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	podList, err := clusterContext.Core.Pod().List(namespace, listOpts)
	if err != nil {
		return nil, err
	}

	return &PodList{Items: podList.Items}, nil
}

// Names returns each Pod name in the list as a new slice of strings.
func (list *PodList) Names() []string {
	var names []string
	for _, p := range list.Items {
		names = append(names, p.Name)
	}
	return names
}
