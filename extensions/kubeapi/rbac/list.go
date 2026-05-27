package rbac

import (
	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RoleList is a struct that contains a list of Roles.
type RoleList struct {
	Items []rbacv1.Role
}

// ClusterRoleList is a struct that contains a list of ClusterRoles.
type ClusterRoleList struct {
	Items []rbacv1.ClusterRole
}

// RoleBindingList is a struct that contains a list of RoleBindings.
type RoleBindingList struct {
	Items []rbacv1.RoleBinding
}

// ClusterRoleBindingList is a struct that contains a list of ClusterRoleBindings.
type ClusterRoleBindingList struct {
	Items []rbacv1.ClusterRoleBinding
}

// ListRoles lists roles in a namespace for a specific cluster.
func ListRoles(client *rancher.Client, clusterID, namespace string, listOpt metav1.ListOptions) (*RoleList, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	roleList, err := clusterContext.RBAC.Role().List(namespace, listOpt)
	if err != nil {
		return nil, err
	}

	return &RoleList{Items: roleList.Items}, nil
}

// ListClusterRoles lists cluster roles for a specific cluster.
func ListClusterRoles(client *rancher.Client, clusterID string, listOpt metav1.ListOptions) (*ClusterRoleList, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	clusterRoleList, err := clusterContext.RBAC.ClusterRole().List(listOpt)
	if err != nil {
		return nil, err
	}

	return &ClusterRoleList{Items: clusterRoleList.Items}, nil
}

// ListRoleBindings lists rolebindings in a namespace for a specific cluster.
func ListRoleBindings(client *rancher.Client, clusterID, namespace string, listOpt metav1.ListOptions) (*RoleBindingList, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	roleBindingList, err := clusterContext.RBAC.RoleBinding().List(namespace, listOpt)
	if err != nil {
		return nil, err
	}

	return &RoleBindingList{Items: roleBindingList.Items}, nil
}

// ListClusterRoleBindings lists clusterrolebindings for a specific cluster.
func ListClusterRoleBindings(client *rancher.Client, clusterID string, listOpt metav1.ListOptions) (*ClusterRoleBindingList, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	clusterRoleBindingList, err := clusterContext.RBAC.ClusterRoleBinding().List(listOpt)
	if err != nil {
		return nil, err
	}

	return &ClusterRoleBindingList{Items: clusterRoleBindingList.Items}, nil
}

// Names returns each Role name in the list as a new slice of strings.
func (list *RoleList) Names() []string {
	var names []string
	for _, r := range list.Items {
		names = append(names, r.Name)
	}
	return names
}

// Names returns each ClusterRole name in the list as a new slice of strings.
func (list *ClusterRoleList) Names() []string {
	var names []string
	for _, r := range list.Items {
		names = append(names, r.Name)
	}
	return names
}

// Names returns each RoleBinding name in the list as a new slice of strings.
func (list *RoleBindingList) Names() []string {
	var names []string
	for _, r := range list.Items {
		names = append(names, r.Name)
	}
	return names
}

// Names returns each ClusterRoleBinding name in the list as a new slice of strings.
func (list *ClusterRoleBindingList) Names() []string {
	var names []string
	for _, r := range list.Items {
		names = append(names, r.Name)
	}
	return names
}
