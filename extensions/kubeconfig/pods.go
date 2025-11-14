package kubeconfig

import (
	"context"
	"errors"

	"github.com/rancher/shepherd/clients/rancher"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	k8Scheme "k8s.io/client-go/kubernetes/scheme"
)

// GetPods utilizes the upstream K8s corev1 client (from the rancher.Client) to get the given cluster's Pods based on any List options passed
func GetPods(client *rancher.Client, clusterID string, namespace string, listOptions *metav1.ListOptions) ([]corev1.Pod, error) {

	kubeConfig, err := GetKubeconfig(client, clusterID)
	if err != nil {
		return nil, err
	}

	restConfig, err := (*kubeConfig).ClientConfig()
	if err != nil {
		return nil, err
	}
	restConfig.ContentConfig.NegotiatedSerializer = serializer.NewCodecFactory(k8Scheme.Scheme)
	restConfig.ContentConfig.GroupVersion = &podGroupVersion
	restConfig.APIPath = apiPath

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), *listOptions)
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

// GetPodNames calls GetPods and filters the list of Pods and extracts their names into a []string
func GetPodNames(client *rancher.Client, clusterID string, namespace string, listOptions *metav1.ListOptions) ([]string, error) {
	pods, err := GetPods(client, clusterID, namespace, listOptions)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, pod := range pods {
		names = append(names, pod.Name)
	}
	if len(names) == 0 {
		return nil, errors.New("GetPodNames: no pod names found")
	}

	return names, nil
}
