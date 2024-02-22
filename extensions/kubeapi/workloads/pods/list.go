package pods

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/api/scheme"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodList is a struct that contains a list of deployments.
type PodList struct {
	Items []corev1.Pod
}

// ListPods is a helper function that uses the dynamic client to list pods on a namespace for a specific cluster with its list options.
func ListPods(client *rancher.Client, clusterID string, namespace string, listOpts metav1.ListOptions) (*PodList, error) {
	podList := new(PodList)

	dynamicClient, err := client.GetDownStreamClusterClient(clusterID)
	if err != nil {
		return nil, err
	}
	podResource := dynamicClient.Resource(PodGroupVersionResource).Namespace(namespace)
	pods, err := podResource.List(context.TODO(), listOpts)
	if err != nil {
		return nil, err
	}

	for _, unstructuredPod := range pods.Items {
		newPod := &corev1.Pod{}
		err := scheme.Scheme.Convert(&unstructuredPod, newPod, unstructuredPod.GroupVersionKind())
		if err != nil {
			return nil, err
		}

		podList.Items = append(podList.Items, *newPod)
	}

	return podList, nil
}

// Names is a method that accepts PodList as a receiver,
// returns each pod name in the list as a new slice of strings.
func (list *PodList) Names() []string {
	var podNames []string

	for _, pod := range list.Items {
		podNames = append(podNames, pod.Name)
	}

	return podNames
}
