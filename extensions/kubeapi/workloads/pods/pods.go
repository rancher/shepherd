package pods

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

const (
	PauseImage       = "registry.k8s.io/pause:3.9"
	DefaultImageName = "nginx"
)

// PodGroupVersion is the required Group Version for accessing pods in a cluster using the dynamic client.
var PodGroupVersionResource = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "pods",
}

// StatusPods is a helper function that uses wrangler context to list pods for a specific cluster with list options.
func StatusPods(client *rancher.Client, clusterID string, listOpts metav1.ListOptions) ([]string, []error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, []error{err}
	}

	pods, err := clusterContext.Core.Pod().List("", listOpts)
	if err != nil {
		return nil, []error{err}
	}

	var podResults []string
	var podErrors []error
	podResults = append(podResults, "pods Status:\n")

	for _, pod := range pods.Items {
		phase := pod.Status.Phase
		if phase == corev1.PodFailed || phase == corev1.PodUnknown {
			podErrors = append(podErrors, fmt.Errorf("ERROR: %s: %s", pod.Name, phase))
		} else {
			podResults = append(podResults, fmt.Sprintf("INFO: %s: %s\n", pod.Name, phase))
		}
	}
	return podResults, podErrors
}

// WaitForPodRunning waits until the specified pod reaches the Running state
func WaitForPodRunning(client *rancher.Client, clusterID, podNamespace, podName string) error {
	return kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(context.Context) (bool, error) {
		pod, err := GetPodByName(client, clusterID, podNamespace, podName)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}

		switch pod.Status.Phase {
		case corev1.PodRunning:
			return true, nil
		case corev1.PodFailed:
			return false, fmt.Errorf("pod %s failed: %s", pod.Name, pod.Status.Message)
		default:
			return false, nil
		}
	})
}

// GetPodByName returns a Pod by name and namespace using wrangler context.
func GetPodByName(client *rancher.Client, clusterID, podNamespace, podName string) (*corev1.Pod, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	pod, err := clusterContext.Core.Pod().Get(podNamespace, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get Pod %s/%s: %w", podNamespace, podName, err)
	}

	return pod, nil
}
