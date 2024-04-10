package users

import (
	"context"

	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/unstructured"
	"github.com/rancher/shepherd/pkg/api/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UpdateUser is a helper function uses the dynamic client to update an existing user
func UpdateUser(client *rancher.Client, existingUser *v3.User, updatedUser *v3.User) (*v3.User, error) {
	adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
	if err != nil {
		return nil, err
	}

	adminDynamicClient, err := adminClient.GetDownStreamClusterClient(LocalCluster)
	if err != nil {
		return nil, err
	}

	userDynamicClient, err := client.GetDownStreamClusterClient(LocalCluster)
	if err != nil {
		return nil, err
	}

	adminUserResource := adminDynamicClient.Resource(UsersGroupVersionResource)
	userResource := userDynamicClient.Resource(UsersGroupVersionResource)

	userUnstructured, err := adminUserResource.Get(context.TODO(), updatedUser.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	currentUser := &v3.User{}
	err = scheme.Scheme.Convert(userUnstructured, currentUser, userUnstructured.GroupVersionKind())
	if err != nil {
		return nil, err
	}

	updatedUser.ObjectMeta.ResourceVersion = currentUser.ObjectMeta.ResourceVersion

	unstructuredResp, err := userResource.Update(context.TODO(), unstructured.MustToUnstructured(updatedUser), metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	newUser := &v3.User{}
	err = scheme.Scheme.Convert(unstructuredResp, newUser, unstructuredResp.GroupVersionKind())
	if err != nil {
		return nil, err
	}
	return newUser, nil
}

// UpdateUserAttributes is a helper function uses the dynamic client to update an existing userattribute
func UpdateUserAttributes(client *rancher.Client, existingUserAttribute *v3.UserAttribute, updatedUserAttribute *v3.UserAttribute) (*v3.UserAttribute, error) {
	adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
	if err != nil {
		return nil, err
	}

	adminDynamicClient, err := adminClient.GetDownStreamClusterClient(LocalCluster)
	if err != nil {
		return nil, err
	}

	userDynamicClient, err := client.GetDownStreamClusterClient(LocalCluster)
	if err != nil {
		return nil, err
	}

	adminUserAttributeResource := adminDynamicClient.Resource(UserAttributesGroupVersionResource)
	userAttributeResource := userDynamicClient.Resource(UserAttributesGroupVersionResource)

	userUnstructured, err := adminUserAttributeResource.Get(context.TODO(), updatedUserAttribute.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	currentUserAttribute := &v3.UserAttribute{}
	err = scheme.Scheme.Convert(userUnstructured, currentUserAttribute, userUnstructured.GroupVersionKind())
	if err != nil {
		return nil, err
	}

	updatedUserAttribute.ObjectMeta.ResourceVersion = currentUserAttribute.ObjectMeta.ResourceVersion

	unstructuredResp, err := userAttributeResource.Update(context.TODO(), unstructured.MustToUnstructured(updatedUserAttribute), metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	newUserAttribute := &v3.UserAttribute{}
	err = scheme.Scheme.Convert(unstructuredResp, newUserAttribute, unstructuredResp.GroupVersionKind())
	if err != nil {
		return nil, err
	}
	return newUserAttribute, nil
}
