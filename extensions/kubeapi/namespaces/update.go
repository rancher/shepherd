package namespaces

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/unstructured"
	"github.com/rancher/shepherd/pkg/api/scheme"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UpdateNamespace is a helper function that uses the dynamic client to update a namespace.
func UpdateNamespace(client *rancher.Client, clusterID string, updatedNamespace *coreV1.Namespace) (*coreV1.Namespace, error) {
	dynamicClient, err := client.GetDownStreamClusterClient(clusterID)
	if err != nil {
		return nil, err
	}

	namespaceResource := dynamicClient.Resource(NamespaceGroupVersionResource).Namespace("")

	namespaceUnstructured, err := namespaceResource.Get(context.TODO(), updatedNamespace.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	currentNamespace := &coreV1.Namespace{}
	err = scheme.Scheme.Convert(namespaceUnstructured, currentNamespace, namespaceUnstructured.GroupVersionKind())
	if err != nil {
		return nil, err
	}

	updatedNamespace.ObjectMeta.ResourceVersion = currentNamespace.ObjectMeta.ResourceVersion

	unstructuredResp, err := namespaceResource.Update(context.TODO(), unstructured.MustToUnstructured(updatedNamespace), metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	newNamespace := &coreV1.Namespace{}
	err = scheme.Scheme.Convert(unstructuredResp, newNamespace, unstructuredResp.GroupVersionKind())
	if err != nil {
		return nil, err
	}

	return newNamespace, nil
}
