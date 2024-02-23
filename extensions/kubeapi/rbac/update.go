package rbac

import (
	"context"

	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/unstructured"
	"github.com/rancher/shepherd/pkg/api/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UpdateGlobalRole is a helper function that uses the dynamic client to update a Global Role
func UpdateGlobalRole(client *rancher.Client, updatedGlobalRole *v3.GlobalRole) (*v3.GlobalRole, error) {
	dynamicClient, err := client.GetDownStreamClusterClient(LocalCluster)
	if err != nil {
		return nil, err
	}
	globalRoleResource := dynamicClient.Resource(GlobalRoleGroupVersionResource)
	globalRolesUnstructured, err := globalRoleResource.Get(context.TODO(), updatedGlobalRole.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	currentGlobalRole := &v3.GlobalRole{}
	err = scheme.Scheme.Convert(globalRolesUnstructured, currentGlobalRole, globalRolesUnstructured.GroupVersionKind())
	if err != nil {
		return nil, err
	}

	updatedGlobalRole.ObjectMeta.ResourceVersion = currentGlobalRole.ObjectMeta.ResourceVersion

	unstructuredResp, err := globalRoleResource.Update(context.TODO(), unstructured.MustToUnstructured(updatedGlobalRole), metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	newGlobalRole := &v3.GlobalRole{}
	err = scheme.Scheme.Convert(unstructuredResp, newGlobalRole, unstructuredResp.GroupVersionKind())
	if err != nil {
		return nil, err
	}
	return newGlobalRole, nil
}

// UpdateRoleTemplate is a helper function that uses the dynamic client to update an existing cluster role template
func UpdateRoleTemplate(client *rancher.Client, updatedCRT *v3.RoleTemplate) (*v3.RoleTemplate, error) {
	dynamicClient, err := client.GetDownStreamClusterClient(LocalCluster)
	if err != nil {
		return nil, err
	}
	crtUnstructured := dynamicClient.Resource(RoleTemplateGroupVersionResource)
	roleTemplate, err := crtUnstructured.Get(context.TODO(), updatedCRT.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	currentCRT := &v3.RoleTemplate{}
	err = scheme.Scheme.Convert(roleTemplate, currentCRT, roleTemplate.GroupVersionKind())
	if err != nil {
		return nil, err
	}

	updatedCRT.ObjectMeta.ResourceVersion = currentCRT.ObjectMeta.ResourceVersion

	unstructuredResp, err := crtUnstructured.Update(context.TODO(), unstructured.MustToUnstructured(updatedCRT), metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	newCRT := &v3.RoleTemplate{}
	err = scheme.Scheme.Convert(unstructuredResp, newCRT, unstructuredResp.GroupVersionKind())
	if err != nil {
		return nil, err
	}
	return newCRT, nil

}

// UpdateClusterRoleTemplateBindings is a helper function that uses the dynamic client to update an existing cluster role template binding
func UpdateClusterRoleTemplateBindings(client *rancher.Client, namespace string, updatedCRTB *v3.ClusterRoleTemplateBinding) (*v3.ClusterRoleTemplateBinding, error) {
	dynamicClient, err := client.GetDownStreamClusterClient(LocalCluster)
	if err != nil {
		return nil, err
	}
	crtbUnstructured := dynamicClient.Resource(ClusterRoleTemplateBindingGroupVersionResource).Namespace(namespace)
	clusterRoleTemplateBinding, err := crtbUnstructured.Get(context.TODO(), updatedCRTB.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	currentCRTB := &v3.ClusterRoleTemplateBinding{}
	err = scheme.Scheme.Convert(clusterRoleTemplateBinding, currentCRTB, clusterRoleTemplateBinding.GroupVersionKind())
	if err != nil {
		return nil, err
	}

	updatedCRTB.ObjectMeta.ResourceVersion = currentCRTB.ObjectMeta.ResourceVersion

	unstructuredResp, err := crtbUnstructured.Update(context.TODO(), unstructured.MustToUnstructured(updatedCRTB), metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	newCRTB := &v3.ClusterRoleTemplateBinding{}
	err = scheme.Scheme.Convert(unstructuredResp, newCRTB, unstructuredResp.GroupVersionKind())
	if err != nil {
		return nil, err
	}
	return newCRTB, nil

}