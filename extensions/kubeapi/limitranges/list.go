package limitranges

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/api/scheme"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// LimitRangeGroupVersionResource is the required Group Version Resource for accessing limit ranges in a cluster,
// using the dynamic client.
var LimitRangeGroupVersionResource = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "limitranges",
}

// LimitRangeList is a struct that contains a list of resource quotas.
type LimitRangeList struct {
	Items []corev1.LimitRange
}

// ListLimitRange is a helper function that uses the dynamic client to list limit range in a cluster with its list options.
func ListLimitRange(client *rancher.Client, clusterID string, namespace string, listOpts metav1.ListOptions) (*LimitRangeList, error) {
	limitRangeList := new(LimitRangeList)

	dynamicClient, err := client.GetDownStreamClusterClient(clusterID)
	if err != nil {
		return nil, err
	}

	limitRangeResource := dynamicClient.Resource(LimitRangeGroupVersionResource).Namespace(namespace)
	limitRange, err := limitRangeResource.List(context.TODO(), listOpts)
	if err != nil {
		return nil, err
	}

	for _, unstructuredLimitRange := range limitRange.Items {
		newlimitRange := &corev1.LimitRange{}
		err := scheme.Scheme.Convert(&unstructuredLimitRange, newlimitRange, unstructuredLimitRange.GroupVersionKind())
		if err != nil {
			return nil, err
		}

		limitRangeList.Items = append(limitRangeList.Items, *newlimitRange)
	}

	return limitRangeList, nil
}
