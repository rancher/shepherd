package kubectl

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/session"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1Unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func ListUnstructured(s *session.Session, client *rancher.Client, name, clusterID, n string, gvr schema.GroupVersionResource, opts metav1.ListOptions) (*v1Unstructured.UnstructuredList, error) {
	dynClient, _, err := setupDynamicClient(s, client, nil, clusterID)
	if err != nil {
		return nil, err
	}

	result, err := dynClient.Resource(gvr).Namespace(n).List(context.TODO(), opts)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func ListAllUnstructured(s *session.Session, client *rancher.Client, name, clusterID string, gvr schema.GroupVersionResource, opts metav1.ListOptions) (*v1Unstructured.UnstructuredList, error) {
	dynClient, _, err := setupDynamicClient(s, client, nil, clusterID)
	if err != nil {
		return nil, err
	}

	result, err := dynClient.Resource(gvr).List(context.TODO(), opts)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func ListUnstructuredFromFlags(s *session.Session, masterURL, kubeconfigPath, name, clusterID, n string, gvr schema.GroupVersionResource, opts metav1.ListOptions) (*v1Unstructured.UnstructuredList, error) {
	dynClient, _, err := setupDynamicClientFromFlags(s, masterURL, kubeconfigPath, nil)
	if err != nil {
		return nil, err
	}

	result, err := dynClient.Resource(gvr).Namespace(n).List(context.TODO(), opts)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func ListAllUnstructuredFromFlags(s *session.Session, masterURL, kubeconfigPath, name, clusterID string, gvr schema.GroupVersionResource, opts metav1.ListOptions) (*v1Unstructured.UnstructuredList, error) {
	dynClient, _, err := setupDynamicClientFromFlags(s, masterURL, kubeconfigPath, nil)
	if err != nil {
		return nil, err
	}

	result, err := dynClient.Resource(gvr).List(context.TODO(), opts)
	if err != nil {
		return nil, err
	}
	return result, nil
}
