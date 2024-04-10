package users

import (
	"context"

	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/api/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// GetUserByID is a helper function that uses the dynamic client to get a specific user from the local cluster.
func GetUserByID(client *rancher.Client, userId string, getOpts metav1.GetOptions) (*v3.User, error) {
	dynamicClient, err := client.GetDownStreamClusterClient(LocalCluster)
	if err != nil {
		return nil, err
	}

	userResource := dynamicClient.Resource(UsersGroupVersionResource)

	unstructuredResp, err := userResource.Get(context.TODO(), userId, getOpts)
	if err != nil {
		return nil, err
	}

	user := &v3.User{}
	err = scheme.Scheme.Convert(unstructuredResp, user, unstructuredResp.GroupVersionKind())
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByAttribute(client *rancher.Client, userId string, getOpts metav1.GetOptions) (*v3.UserAttribute, error) {
	dynamicClient, err := client.GetDownStreamClusterClient(LocalCluster)
	if err != nil {
		return nil, err
	}

	userResource := dynamicClient.Resource(UserAttributesGroupVersionResource)

	unstructuredResp, err := userResource.Get(context.TODO(), userId, getOpts)
	if err != nil {
		return nil, err
	}

	userAttribute := &v3.UserAttribute{}
	err = scheme.Scheme.Convert(unstructuredResp, userAttribute, unstructuredResp.GroupVersionKind())
	if err != nil {
		return nil, err
	}
	return userAttribute, nil
}
