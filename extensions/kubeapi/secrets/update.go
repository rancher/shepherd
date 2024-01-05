package secrets

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/unstructured"
	"github.com/rancher/shepherd/pkg/api/scheme"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UpdateSecret is a helper function that uses the dynamic client to update a secret in a specific cluster.
func UpdateSecret(client *rancher.Client, clusterID string, existingSecret *corev1.Secret, updatedSecret *corev1.Secret) (*corev1.Secret, error) {
	dynamicClient, err := client.GetDownStreamClusterClient(clusterID)
	if err != nil {
		return nil, err
	}

	secretResource := dynamicClient.Resource(SecretGroupVersionResource).Namespace(existingSecret.Namespace)

	updatedSecret.ObjectMeta.ResourceVersion = existingSecret.ObjectMeta.ResourceVersion

	unstructuredResp, err := secretResource.Update(context.TODO(), unstructured.MustToUnstructured(updatedSecret), metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	newSecret := &corev1.Secret{}
	err = scheme.Scheme.Convert(unstructuredResp, newSecret, unstructuredResp.GroupVersionKind())
	if err != nil {
		return nil, err
	}

	return newSecret, nil
}
