package users

import (
	"context"

	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/api/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListUsers is a helper function that returns the user ID by name using dynamic client
func ListUsers(client *rancher.Client, listOpt metav1.ListOptions) (*v3.UserList, error) {
	dynamicClient, err := client.GetDownStreamClusterClient(LocalCluster)
	if err != nil {
		return nil, err
	}

	unstructuredList, err := dynamicClient.Resource(UsersGroupVersionResource).List(context.TODO(), listOpt)
	if err != nil {
		return nil, err
	}

	userList := new(v3.UserList)
	for _, unstructuredRB := range unstructuredList.Items {
		user := &v3.User{}
		err := scheme.Scheme.Convert(&unstructuredRB, user, unstructuredRB.GroupVersionKind())
		if err != nil {
			return nil, err
		}

		userList.Items = append(userList.Items, *user)
	}

	return userList, nil

}

// ListUserAttributes is a helper function that uses the dynamic client to list user attributes from local cluster.
func ListUserAttributes(client *rancher.Client, listOpt metav1.ListOptions) (*v3.UserAttributeList, error) {
	dynamicClient, err := client.GetDownStreamClusterClient(LocalCluster)
	if err != nil {
		return nil, err
	}

	unstructuredList, err := dynamicClient.Resource(UserAttributesGroupVersionResource).List(context.Background(), listOpt)
	if err != nil {
		return nil, err
	}

	userAttributesList := new(v3.UserAttributeList)
	for _, unstructuredRB := range unstructuredList.Items {
		userAttributes := &v3.UserAttribute{}
		err := scheme.Scheme.Convert(&unstructuredRB, userAttributes, unstructuredRB.GroupVersionKind())
		if err != nil {
			return nil, err
		}

		userAttributesList.Items = append(userAttributesList.Items, *userAttributes)
	}

	return userAttributesList, nil
}
