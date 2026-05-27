package daemonsets

import (
	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DaemonSetList is a struct that contains a list of DaemonSets.
type DaemonSetList struct {
	Items []appsv1.DaemonSet
}

// ListDaemonSets returns DaemonSets in a specific namespace using wrangler context and returns a DaemonSetList.
func ListDaemonSets(client *rancher.Client, clusterID, daemonSetNamespace string, listOpts metav1.ListOptions) (*DaemonSetList, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	daemonSetList, err := clusterContext.Apps.DaemonSet().List(daemonSetNamespace, listOpts)
	if err != nil {
		return nil, err
	}

	return &DaemonSetList{Items: daemonSetList.Items}, nil
}

// Names returns each DaemonSet name in the list as a new slice of strings.
func (list *DaemonSetList) Names() []string {
	var names []string
	for _, ds := range list.Items {
		names = append(names, ds.Name)
	}
	return names
}
