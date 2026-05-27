package statefulsets

import (
	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatefulSetList is a struct that contains a list of StatefulSets.
type StatefulSetList struct {
	Items []appsv1.StatefulSet
}

// ListStatefulSets returns StatefulSets in a specific namespace using wrangler context and returns a StatefulSetList.
func ListStatefulSets(client *rancher.Client, clusterID, namespace string, listOpts metav1.ListOptions) (*StatefulSetList, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	ssList, err := clusterContext.Apps.StatefulSet().List(namespace, listOpts)
	if err != nil {
		return nil, err
	}

	return &StatefulSetList{Items: ssList.Items}, nil
}

// Names returns each StatefulSet name in the list as a new slice of strings.
func (list *StatefulSetList) Names() []string {
	var names []string
	for _, ss := range list.Items {
		names = append(names, ss.Name)
	}
	return names
}
