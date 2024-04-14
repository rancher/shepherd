package users

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	GroupName    = "management.cattle.io"
	Version      = "v3"
	LocalCluster = "local"
)

// UserAttributesGroupVersionResource is the required Group Version Resource for accessing user attributes in local cluster, using the dynamic client.
var UserAttributesGroupVersionResource = schema.GroupVersionResource{
	Group:    GroupName,
	Version:  Version,
	Resource: "userattributes",
}

// UsersGroupVersionResource is the required Group Version Resource for accessing user attributes in local cluster, using the dynamic client.
var UsersGroupVersionResource = schema.GroupVersionResource{
	Group:    GroupName,
	Version:  Version,
	Resource: "users",
}
