package pods

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/api/scheme"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatusPods is a helper function that uses the dynamic client to list pods on a namespace for a specific cluster with its list options.
func StatusPods(client *rancher.Client, clusterID string, listOpts metav1.ListOptions) ([]string, []error) {
	var podList []corev1.Pod

	dynamicClient, err := client.GetDownStreamClusterClient(clusterID)
	if err != nil {
		return nil, []error{err}
	}
	podResource := dynamicClient.Resource(PodGroupVersionResource)
	pods, err := podResource.List(context.TODO(), listOpts)
	if err != nil {
		return nil, []error{err}
	}

	for _, unstructuredPod := range pods.Items {
		newPod := &corev1.Pod{}
		err := scheme.Scheme.Convert(&unstructuredPod, newPod, unstructuredPod.GroupVersionKind())
		if err != nil {
			return nil, []error{err}
		}

		podList = append(podList, *newPod)
	}

	var podResults []string
	var podErrors []error
	podResults = append(podResults, "pods Status:\n")

	for _, pod := range podList {
		phase := pod.Status.Phase
		if phase == corev1.PodFailed || phase == corev1.PodUnknown {
			podErrors = append(podErrors, fmt.Errorf("ERROR: %s: %s", pod.Name, phase))
		} else {
			podResults = append(podResults, fmt.Sprintf("INFO: %s: %s\n", pod.Name, phase))
		}
	}
	return podResults, podErrors
}
