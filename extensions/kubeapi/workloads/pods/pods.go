package pods

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// PodGroupVersion is the required Group Version for accessing pods in a cluster,
// using the dynamic client.
var PodGroupVersionResource = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "pods",
}
