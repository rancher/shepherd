package rbac

import (
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeleteRole deletes a role by name from a specific cluster
func DeleteRole(client *rancher.Client, clusterID, namespace, roleName string) error {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return err
	}

	err = clusterContext.RBAC.Role().Delete(namespace, roleName, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}

// DeleteClusterRole deletes a cluster role by name from a specific cluster
func DeleteClusterRole(client *rancher.Client, clusterID, clusterRoleName string) error {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return err
	}

	err = clusterContext.RBAC.ClusterRole().Delete(clusterRoleName, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}

// DeleteRoleBinding deletes a role binding by name from a specific cluster
func DeleteRoleBinding(client *rancher.Client, clusterID, namespace, roleBindingName string) error {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return err
	}

	err = clusterContext.RBAC.RoleBinding().Delete(namespace, roleBindingName, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}

// DeleteClusterRoleBinding deletes a cluster role binding by name from a specific cluster
func DeleteClusterRoleBinding(client *rancher.Client, clusterID, clusterRoleBindingName string) error {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return err
	}

	err = clusterContext.RBAC.ClusterRoleBinding().Delete(clusterRoleBindingName, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}

// DeleteGlobalRole deletes a Global Role by name using wrangler context
func DeleteGlobalRole(client *rancher.Client, globalRoleName string, waitForDelete bool) error {
	err := client.WranglerContext.Mgmt.GlobalRole().Delete(globalRoleName, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if waitForDelete {
		err = WaitForGlobalRoleDeletion(client, globalRoleName)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteClusterRoleTemplateBinding deletes the cluster role template binding by name using wrangler context
func DeleteClusterRoleTemplateBinding(client *rancher.Client, crtbNamespace, crtbName string, waitForDelete bool) error {
	err := client.WranglerContext.Mgmt.ClusterRoleTemplateBinding().Delete(crtbNamespace, crtbName, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if waitForDelete {
		err = WaitForClusterRoleTemplateBindingDeletion(client, crtbNamespace, crtbName)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteProjectRoleTemplateBinding deletes the project role template binding by name using wrangler context
func DeleteProjectRoleTemplateBinding(client *rancher.Client, prtbNamespace, prtbName string, waitForDelete bool) error {
	err := client.WranglerContext.Mgmt.ProjectRoleTemplateBinding().Delete(prtbNamespace, prtbName, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if waitForDelete {
		err = WaitForProjectRoleTemplateBindingDeletion(client, prtbNamespace, prtbName)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteRoleTemplate deletes a role template by name using wrangler context
func DeleteRoleTemplate(client *rancher.Client, roleTemplateName string, waitForDelete bool) error {
	err := client.WranglerContext.Mgmt.RoleTemplate().Delete(roleTemplateName, nil)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if waitForDelete {
		err = WaitForRoleTemplateDeletion(client, roleTemplateName)
		if err != nil {
			return fmt.Errorf("role template %s not deleted in time: %w", roleTemplateName, err)
		}
	}

	return nil
}

// DeleteGlobalRoleBinding deletes a global role binding by name using wrangler context
func DeleteGlobalRoleBinding(client *rancher.Client, globalRoleBindingName string, waitForDelete bool) error {
	err := client.WranglerContext.Mgmt.GlobalRoleBinding().Delete(globalRoleBindingName, nil)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}

		return fmt.Errorf("failed to delete global role binding %s: %w", globalRoleBindingName, err)
	}

	if waitForDelete {
		err = WaitForGlobalRoleBindingDeletion(client, globalRoleBindingName)
		if err != nil {
			return fmt.Errorf("global role binding %s not deleted in time: %w", globalRoleBindingName, err)
		}
	}

	return nil
}
