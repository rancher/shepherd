package rbac

import (
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	rbacv1 "k8s.io/api/rbac/v1"
)

// CreateRole is a helper function that uses the wrangler context to create a role on a namespace for a specific cluster.
func CreateRole(client *rancher.Client, clusterID string, role *rbacv1.Role) (*rbacv1.Role, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	newRole, err := clusterContext.RBAC.Role().Create(role)
	if err != nil {
		return nil, err
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
			if err != nil {
				return err
			}

			return DeleteRole(adminClient, clusterID, newRole.Namespace, newRole.Name)
		})
	}

	return newRole, nil
}

// CreateClusterRole is a helper function that uses the wrangler context to create a cluster role for a specific cluster.
func CreateClusterRole(client *rancher.Client, clusterID string, clusterRole *rbacv1.ClusterRole) (*rbacv1.ClusterRole, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	newClusterRole, err := clusterContext.RBAC.ClusterRole().Create(clusterRole)
	if err != nil {
		return nil, err
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
			if err != nil {
				return err
			}

			return DeleteClusterRole(adminClient, clusterID, newClusterRole.Name)
		})
	}

	return newClusterRole, nil
}

// CreateRoleBinding is a helper function that uses the wrangler context to create a rolebinding on a namespace for a specific cluster.
func CreateRoleBinding(client *rancher.Client, clusterID string, roleBinding *rbacv1.RoleBinding) (*rbacv1.RoleBinding, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	newRoleBinding, err := clusterContext.RBAC.RoleBinding().Create(roleBinding)
	if err != nil {
		return nil, err
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
			if err != nil {
				return err
			}

			return DeleteRoleBinding(adminClient, clusterID, newRoleBinding.Namespace, newRoleBinding.Name)
		})
	}

	return newRoleBinding, nil
}

// CreateClusterRoleBinding is a helper function that uses the wrangler context to create a clusterrolebinding for a specific cluster.
func CreateClusterRoleBinding(client *rancher.Client, clusterID string, clusterRoleBinding *rbacv1.ClusterRoleBinding) (*rbacv1.ClusterRoleBinding, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	newClusterRoleBinding, err := clusterContext.RBAC.ClusterRoleBinding().Create(clusterRoleBinding)
	if err != nil {
		return nil, err
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
			if err != nil {
				return err
			}

			return DeleteClusterRoleBinding(adminClient, clusterID, newClusterRoleBinding.Name)
		})
	}

	return newClusterRoleBinding, nil
}

// CreateGlobalRole is a helper function that uses wrangler context to create a Global Role
func CreateGlobalRole(client *rancher.Client, globalRole *v3.GlobalRole) (*v3.GlobalRole, error) {
	newGlobalRole, err := client.WranglerContext.Mgmt.GlobalRole().Create(globalRole)
	if err != nil {
		return nil, err
	}

	latestGlobalRole, err := WaitForGlobalRoleToExist(client, newGlobalRole.Name)
	if err != nil {
		return nil, err
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
			if err != nil {
				return err
			}

			return DeleteGlobalRole(adminClient, latestGlobalRole.Name, true)
		})
	}

	return latestGlobalRole, nil
}

// CreateGlobalRoleBinding is a helper function that uses wrangler context to create a Global Role Binding
func CreateGlobalRoleBinding(client *rancher.Client, globalRoleBinding *v3.GlobalRoleBinding) (*v3.GlobalRoleBinding, error) {
	newGlobalRoleBinding, err := client.WranglerContext.Mgmt.GlobalRoleBinding().Create(globalRoleBinding)
	if err != nil {
		return nil, err
	}

	latestGlobalRoleBinding, err := WaitForGlobalRoleBindingToExist(client, newGlobalRoleBinding.Name)
	if err != nil {
		return nil, err
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
			if err != nil {
				return err
			}

			return DeleteGlobalRoleBinding(adminClient, latestGlobalRoleBinding.Name, true)
		})
	}

	return latestGlobalRoleBinding, nil
}

// CreateRoleTemplate creates a cluster or project role template with the provided rules using wrangler context
func CreateRoleTemplate(client *rancher.Client, roleTemplate *v3.RoleTemplate) (*v3.RoleTemplate, error) {
	newRoleTemplate, err := client.WranglerContext.Mgmt.RoleTemplate().Create(roleTemplate)
	if err != nil {
		return nil, err
	}

	latestRoleTemplate, err := WaitForRoleTemplateToExist(client, newRoleTemplate.Name)
	if err != nil {
		return nil, err
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
			if err != nil {
				return err
			}

			return DeleteRoleTemplate(adminClient, latestRoleTemplate.Name, true)
		})
	}

	return latestRoleTemplate, nil
}

// CreateClusterRoleTemplateBinding creates a cluster role template binding for the user with the provided role template using wrangler context
func CreateClusterRoleTemplateBinding(client *rancher.Client, clusterRoleTemplateBinding *v3.ClusterRoleTemplateBinding) (*v3.ClusterRoleTemplateBinding, error) {
	newCRTB, err := client.WranglerContext.Mgmt.ClusterRoleTemplateBinding().Create(clusterRoleTemplateBinding)
	if err != nil {
		return nil, err
	}

	latestCRTB, err := WaitForClusterRoleTemplateBindingToExist(client, newCRTB.Namespace, newCRTB.Name)
	if err != nil {
		return nil, err
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
			if err != nil {
				return err
			}

			return DeleteClusterRoleTemplateBinding(adminClient, latestCRTB.Namespace, latestCRTB.Name, true)
		})
	}

	return latestCRTB, nil
}

// CreateProjectRoleTemplateBinding creates a project role template binding for the user with the provided role template using wrangler context
func CreateProjectRoleTemplateBinding(client *rancher.Client, projectRoleTemplateBinding *v3.ProjectRoleTemplateBinding) (*v3.ProjectRoleTemplateBinding, error) {
	newPRTB, err := client.WranglerContext.Mgmt.ProjectRoleTemplateBinding().Create(projectRoleTemplateBinding)
	if err != nil {
		return nil, err
	}

	latestPRTB, err := WaitForProjectRoleTemplateBindingToExist(client, newPRTB.ProjectName, newPRTB.Namespace, newPRTB.Name, newPRTB.UserName)
	if err != nil {
		return nil, err
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
			if err != nil {
				return err
			}

			return DeleteProjectRoleTemplateBinding(adminClient, latestPRTB.Namespace, latestPRTB.Name, true)
		})
	}

	return latestPRTB, nil
}
