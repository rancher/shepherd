package configmaps

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/unstructured"
	"github.com/rancher/shepherd/pkg/api/scheme"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type PatchOP string

const (
	AddPatchOP     PatchOP = "add"
	ReplacePatchOP PatchOP = "replace"
	RemovePatchOP  PatchOP = "remove"
)

// ConfigMapGroupVersionResource is the required Group Version Resource for accessing config maps in a cluster,
// using the dynamic client.
var ConfigMapGroupVersionResource = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "configmaps",
}

type ConfigMapList struct {
	Items []coreV1.ConfigMap
}

// CreateConfigMap is a helper function that uses the dynamic client to create a config map on a namespace for a specific cluster.
// It registers a delete fuction.
func CreateConfigMap(client *rancher.Client, clusterID, configMapName, description, namespace string, data, labels, annotations map[string]string) (*coreV1.ConfigMap, error) {
	// ConfigMap object for a namespace in a cluster
	annotations["field.cattle.io/description"] = description
	configMap := &coreV1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        configMapName,
			Annotations: annotations,
			Namespace:   namespace,
			Labels:      labels,
		},
		Data: data,
	}

	dynamicClient, err := client.GetDownStreamClusterClient(clusterID)
	if err != nil {
		return nil, err
	}

	configMapResource := dynamicClient.Resource(ConfigMapGroupVersionResource).Namespace(namespace)

	unstructuredResp, err := configMapResource.Create(context.TODO(), unstructured.MustToUnstructured(configMap), metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	newConfig := &coreV1.ConfigMap{}
	err = scheme.Scheme.Convert(unstructuredResp, newConfig, unstructuredResp.GroupVersionKind())
	if err != nil {
		return nil, err
	}
	return newConfig, nil
}

// NewConfigmapTemplate is a constructor that creates a configmap template
func NewConfigmapTemplate(configmapName, namespace string, annotations, labels, data map[string]string) coreV1.ConfigMap {
	if annotations == nil {
		annotations = make(map[string]string)
	}
	if labels == nil {
		labels = make(map[string]string)
	}
	if data == nil {
		data = make(map[string]string)
	}

	return coreV1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        configmapName,
			Namespace:   namespace,
			Annotations: annotations,
			Labels:      labels,
		},
		Data: data,
	}
}

func ListConfigMaps(client *rancher.Client, clusterID, namespace string, opts metav1.ListOptions) (*ConfigMapList, error) {
	configMapList := new(ConfigMapList)

	dynamicClient, err := client.GetDownStreamClusterClient(clusterID)
	if err != nil {
		return nil, err
	}

	configMapResource := dynamicClient.Resource(ConfigMapGroupVersionResource).Namespace(namespace)
	configMaps, err := configMapResource.List(context.TODO(), opts)
	if err != nil {
		return nil, err
	}

	for _, unstructuredConfigMap := range configMaps.Items {
		newConfigMap := &coreV1.ConfigMap{}

		err := scheme.Scheme.Convert(&unstructuredConfigMap, newConfigMap, unstructuredConfigMap.GroupVersionKind())
		if err != nil {
			return nil, err
		}

		configMapList.Items = append(configMapList.Items, *newConfigMap)
	}

	return configMapList, nil
}

func GetConfigMapByName(client *rancher.Client, clusterID, configMapName, namespace string, getOpts metav1.GetOptions) (*coreV1.ConfigMap, error) {
	dynamicClient, err := client.GetDownStreamClusterClient(clusterID)
	if err != nil {
		return nil, err
	}

	configMapResource := dynamicClient.Resource(ConfigMapGroupVersionResource).Namespace(namespace)
	unstructuredResp, err := configMapResource.Get(context.TODO(), configMapName, getOpts)
	if err != nil {
		return nil, err
	}

	newConfigMap := &coreV1.ConfigMap{}
	err = scheme.Scheme.Convert(unstructuredResp, newConfigMap, unstructuredResp.GroupVersionKind())
	if err != nil {
		return nil, err
	}
	return newConfigMap, nil
}

func PatchConfigMap(client *rancher.Client, clusterID, configMapName, namespace string, data string, patchType types.PatchType) (*coreV1.ConfigMap, error) {
	dynamicClient, err := client.GetDownStreamClusterClient(clusterID)
	if err != nil {
		return nil, err
	}
	configMapResource := dynamicClient.Resource(ConfigMapGroupVersionResource).Namespace(namespace)

	unstructuredResp, err := configMapResource.Patch(context.TODO(), configMapName, patchType, []byte(data), metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}

	newConfigMap := &coreV1.ConfigMap{}
	err = scheme.Scheme.Convert(unstructuredResp, newConfigMap, unstructuredResp.GroupVersionKind())
	if err != nil {
		return nil, err
	}
	return newConfigMap, nil
}

// PatchConfigMapFromYAML is a helper function that uses the dynamic client to patch a configMap in a namespace for a specific cluster.
// Different merge strategies are supported based on the PatchType.
func PatchConfigMapFromYAML(client *rancher.Client, clusterID, configMapName, namespace string, rawYAML []byte, patchType types.PatchType) (*coreV1.ConfigMap, error) {
	dynamicClient, err := client.GetDownStreamClusterClient(clusterID)
	if err != nil {
		return nil, err
	}
	configMapResource := dynamicClient.Resource(ConfigMapGroupVersionResource).Namespace(namespace)

	rawJSON, err := yaml.ToJSON(rawYAML)
	if err != nil {
		return nil, err
	}

	unstructuredResp, err := configMapResource.Patch(context.TODO(), configMapName, patchType, rawJSON, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}

	newConfigMap := &coreV1.ConfigMap{}
	err = scheme.Scheme.Convert(unstructuredResp, newConfigMap, unstructuredResp.GroupVersionKind())
	if err != nil {
		return nil, err
	}
	return newConfigMap, nil
}
