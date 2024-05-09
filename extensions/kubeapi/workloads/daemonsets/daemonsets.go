package daemonsets

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults/schema/groupversionresources"
	"github.com/rancher/shepherd/pkg/api/scheme"
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetDaemonsetByName is a helper function that uses the dynamic client to get a specific daemonset on a namespace for a specific cluster.
func GetDaemonsetByName(client *rancher.Client, clusterID, namespace, daemonsetName string) (*appv1.DaemonSet, error) {
	dynamicClient, err := client.GetDownStreamClusterClient(clusterID)
	if err != nil {
		return nil, err
	}

	daemonsetResource := dynamicClient.Resource(groupversionresources.Daemonset()).Namespace(namespace)
	unstructuredResp, err := daemonsetResource.Get(context.TODO(), daemonsetName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	newDaemonset := &appv1.DaemonSet{}
	err = scheme.Scheme.Convert(unstructuredResp, newDaemonset, unstructuredResp.GroupVersionKind())
	if err != nil {
		return nil, err
	}

	return newDaemonset, nil
}
