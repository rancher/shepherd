package kubectl

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/session"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1Unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func GetUnstructured(s *session.Session, client *rancher.Client, name, clusterID, n string, gvr schema.GroupVersionResource) (*v1Unstructured.Unstructured, error) {
	dynClient, _, err := setupDynamicClient(s, client, nil, clusterID)
	if err != nil {
		return nil, err
	}

	result, err := dynClient.Resource(gvr).Namespace(n).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetUnstructuredFromFlags(s *session.Session, masterURL, kubeconfigPath, name, clusterID, n string, gvr schema.GroupVersionResource) (*v1Unstructured.Unstructured, error) {
	dynClient, _, err := setupDynamicClientFromFlags(s, masterURL, kubeconfigPath, nil)
	if err != nil {
		return nil, err
	}

	result, err := dynClient.Resource(gvr).Namespace(n).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return result, nil
}
