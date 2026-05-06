package rbac

import (
	"context"
	"fmt"

	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// WaitForGlobalRoleToExist waits until the GlobalRole exists and returns the latest object
func WaitForGlobalRoleToExist(client *rancher.Client, globalRoleName string) (*v3.GlobalRole, error) {
	var globalRole *v3.GlobalRole
	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		grObj, err := client.WranglerContext.Mgmt.GlobalRole().Get(globalRoleName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		globalRole = grObj
		return true, nil
	})
	if err != nil {
		return nil, fmt.Errorf("timed out waiting for GlobalRole %s to exist: %w", globalRoleName, err)
	}

	return globalRole, nil
}

// WaitForGlobalRoleBindingToExist waits until the GlobalRoleBinding exists and returns the latest object
func WaitForGlobalRoleBindingToExist(client *rancher.Client, grbName string) (*v3.GlobalRoleBinding, error) {
	var grb *v3.GlobalRoleBinding
	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		grbObj, err := client.WranglerContext.Mgmt.GlobalRoleBinding().Get(grbName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		grb = grbObj
		return true, nil
	})
	if err != nil {
		return nil, fmt.Errorf("timed out waiting for GlobalRoleBinding %s to exist: %w", grbName, err)
	}

	return grb, nil
}

// WaitForRoleTemplateToExist waits until the RoleTemplate exists and returns the latest object
func WaitForRoleTemplateToExist(client *rancher.Client, roleTemplateName string) (*v3.RoleTemplate, error) {
	var roleTemplate *v3.RoleTemplate
	err := kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		rtObj, err := client.WranglerContext.Mgmt.RoleTemplate().Get(roleTemplateName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		roleTemplate = rtObj
		return true, nil
	})
	if err != nil {
		return nil, fmt.Errorf("timed out waiting for RoleTemplate %s to exist: %w", roleTemplateName, err)
	}

	return roleTemplate, nil
}

// WaitForClusterRoleTemplateBindingToExist waits for the CRTB to reach the Completed status or checks for its existence if status field is not supported (older Rancher versions)
func WaitForClusterRoleTemplateBindingToExist(client *rancher.Client, crtbNamespace, crtbName string) (*v3.ClusterRoleTemplateBinding, error) {
	var crtb *v3.ClusterRoleTemplateBinding
	err := kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
		crtbObj, err := client.WranglerContext.Mgmt.ClusterRoleTemplateBinding().Get(crtbNamespace, crtbName, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		crtb = crtbObj

		if crtb.Status.Summary == "Completed" {
			return true, nil
		}

		if crtb != nil && crtb.Name == crtbName && crtb.Namespace == crtbNamespace {
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return nil, fmt.Errorf("timed out waiting for ClusterRoleTemplateBinding %s/%s to exist: %w", crtbNamespace, crtbName, err)
	}

	return crtb, nil
}

// WaitForProjectRoleTemplateBindingToExist waits for the PRTB to exist with the correct user and project
func WaitForProjectRoleTemplateBindingToExist(client *rancher.Client, projectName, prtbNamespace, prtbName, userID string) (*v3.ProjectRoleTemplateBinding, error) {
	var prtb *v3.ProjectRoleTemplateBinding
	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		prtbObj, err := client.WranglerContext.Mgmt.ProjectRoleTemplateBinding().Get(prtbNamespace, prtbName, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		if prtbObj != nil && prtbObj.UserName == userID && prtbObj.ProjectName == projectName {
			prtb = prtbObj
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return nil, fmt.Errorf("timed out waiting for ProjectRoleTemplateBinding %s/%s (user: %s, project: %s) to exist: %w", prtbNamespace, prtbName, userID, projectName, err)
	}

	return prtb, nil
}

// WaitForClusterRoleBindingToExist waits until the ClusterRoleBinding exists on the cluster and returns the latest object
func WaitForClusterRoleBindingToExist(client *rancher.Client, clusterID, crbName string) (*rbacv1.ClusterRoleBinding, error) {
	var crb *rbacv1.ClusterRoleBinding
	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
		crbObj, err := GetClusterRoleBindingByName(client, clusterID, crbName)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		crb = crbObj
		return true, nil
	})
	if err != nil {
		return nil, fmt.Errorf("timed out waiting for ClusterRoleBinding %s to exist in cluster %s: %w", crbName, clusterID, err)
	}

	return crb, nil
}

// WaitForRoleBindingToExist waits until the RoleBinding exists on the cluster and returns the latest object
func WaitForRoleBindingToExist(client *rancher.Client, clusterID, rbNamespace, rbName string) (*rbacv1.RoleBinding, error) {
	var rb *rbacv1.RoleBinding
	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
		rbObj, err := GetRoleBindingByName(client, clusterID, rbNamespace, rbName)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		rb = rbObj
		return true, nil
	})
	if err != nil {
		return nil, fmt.Errorf("timed out waiting for RoleBinding %s/%s to exist in cluster %s: %w", rbNamespace, rbName, clusterID, err)
	}

	return rb, nil
}

// WaitForGlobalRoleDeletion waits until the GlobalRole is deleted
func WaitForGlobalRoleDeletion(client *rancher.Client, globalRoleName string) error {
	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
		_, err = client.WranglerContext.Mgmt.GlobalRole().Get(globalRoleName, metav1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("timed out waiting for GlobalRole %s to be deleted: %w", globalRoleName, err)
	}

	return nil
}

// WaitForGlobalRoleBindingDeletion waits until the GlobalRoleBinding is deleted
func WaitForGlobalRoleBindingDeletion(client *rancher.Client, globalRoleBindingName string) error {
	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
		_, err = client.WranglerContext.Mgmt.GlobalRoleBinding().Get(globalRoleBindingName, metav1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("timed out waiting for GlobalRoleBinding %s to be deleted: %w", globalRoleBindingName, err)
	}

	return nil
}

// WaitForRoleTemplateDeletion waits until the RoleTemplate is deleted
func WaitForRoleTemplateDeletion(client *rancher.Client, rtName string) error {
	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
		_, err = client.WranglerContext.Mgmt.RoleTemplate().Get(rtName, metav1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("timed out waiting for RoleTemplate %s to be deleted: %w", rtName, err)
	}

	return nil
}

// WaitForClusterRoleTemplateBindingDeletion waits until the ClusterRoleTemplateBinding is deleted
func WaitForClusterRoleTemplateBindingDeletion(client *rancher.Client, crtbNamespace, crtbName string) error {
	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
		_, err = client.WranglerContext.Mgmt.ClusterRoleTemplateBinding().Get(crtbNamespace, crtbName, metav1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("timed out waiting for ClusterRoleTemplateBinding %s/%s to be deleted: %w", crtbNamespace, crtbName, err)
	}

	return nil
}

// WaitForProjectRoleTemplateBindingDeletion waits until the ProjectRoleTemplateBinding is deleted
func WaitForProjectRoleTemplateBindingDeletion(client *rancher.Client, prtbNamespace, prtbName string) error {
	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
		_, err = client.WranglerContext.Mgmt.ProjectRoleTemplateBinding().Get(prtbNamespace, prtbName, metav1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("timed out waiting for ProjectRoleTemplateBinding %s/%s to be deleted: %w", prtbNamespace, prtbName, err)
	}

	return nil
}

// WaitForClusterRoleBindingDeletion waits until the ClusterRoleBinding is deleted from a cluster
func WaitForClusterRoleBindingDeletion(client *rancher.Client, clusterID, crbName string) error {
	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
		_, err = GetClusterRoleBindingByName(client, clusterID, crbName)
		if k8serrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("timed out waiting for ClusterRoleBinding %s to be deleted in cluster %s: %w", crbName, clusterID, err)
	}

	return nil
}

// WaitForRoleBindingDeletion waits until the RoleBinding is deleted from a cluster
func WaitForRoleBindingDeletion(client *rancher.Client, clusterID, rbNamespace, rbName string) error {
	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
		_, err = GetRoleBindingByName(client, clusterID, rbNamespace, rbName)
		if k8serrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("timed out waiting for RoleBinding %s/%s to be deleted in cluster %s: %w", rbNamespace, rbName, clusterID, err)
	}

	return nil
}

// GetRoleByName is a helper function that uses the wrangler context to get a Role by name from a specific cluster and namespace
func GetRoleByName(client *rancher.Client, clusterID, namespaceName, roleName string) (*rbacv1.Role, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	return clusterContext.RBAC.Role().Get(namespaceName, roleName, metav1.GetOptions{})
}

// GetClusterRole is a helper function that uses the wrangler context to get a ClusterRole by name from a specific cluster
func GetClusterRole(client *rancher.Client, clusterID, clusterRoleName string) (*rbacv1.ClusterRole, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	return clusterContext.RBAC.ClusterRole().Get(clusterRoleName, metav1.GetOptions{})
}

// GetRoleBindingByName is a helper function that uses the wrangler context to get a RoleBinding by name from a specific cluster and namespace
func GetRoleBindingByName(client *rancher.Client, clusterID, namespaceName, roleBindingName string) (*rbacv1.RoleBinding, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	return clusterContext.RBAC.RoleBinding().Get(namespaceName, roleBindingName, metav1.GetOptions{})
}

// GetClusterRoleBindingByName is a helper function that uses the wrangler context to get a ClusterRoleBinding by name from a specific cluster
func GetClusterRoleBindingByName(client *rancher.Client, clusterID, clusterRoleBindingName string) (*rbacv1.ClusterRoleBinding, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	return clusterContext.RBAC.ClusterRoleBinding().Get(clusterRoleBindingName, metav1.GetOptions{})
}
