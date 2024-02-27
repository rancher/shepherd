package cluster

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const ManagementGroupName = "management.cattle.io"

func ManagementClusterGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: ManagementGroupName, Version: "v3", Resource: "clusters"}
}

// ListAll is a helper function that uses the dynamic client to return a list of all clusters.management.cattle.io.
func ListAll(client *rancher.Client, opts *metav1.ListOptions) (list *unstructured.UnstructuredList, err error) {
	if opts == nil {
		opts = &metav1.ListOptions{}
	}

	dynamic, err := client.GetRancherDynamicClient()
	if err != nil {
		return
	}

	unstructuredClusterList, err := dynamic.Resource(ManagementClusterGVR()).List(context.TODO(), *opts)
	if err != nil {
		return
	}
	return unstructuredClusterList, err
}
