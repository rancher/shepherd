package projects

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	Admin               = "admin"
	StandardUser        = "user"
	DefaultNamespace    = "fleet-default"
	RancherNamespace    = "cattle-system"
	LocalCluster        = "local"
	Projects            = "projects"
	ProjectIDAnnotation = "field.cattle.io/projectId"
	GroupName           = "management.cattle.io"
	Version             = "v3"
)

// ProjectGroupVersionResource is the required Group Version Resource for accessing projects in a cluster, using the dynamic client.
var ProjectGroupVersionResource = schema.GroupVersionResource{
	Group:    GroupName,
	Version:  Version,
	Resource: Projects,
}
