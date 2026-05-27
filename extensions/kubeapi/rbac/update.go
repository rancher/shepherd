package rbac

import (
	"context"
	"fmt"

	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// UpdateGlobalRole is a helper function that uses wrangler context to update an existing global role
func UpdateGlobalRole(client *rancher.Client, updatedGlobalRole *v3.GlobalRole) (*v3.GlobalRole, error) {
	var updated *v3.GlobalRole
	var lastErr error

	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
		current, getErr := client.WranglerContext.Mgmt.GlobalRole().Get(updatedGlobalRole.Name, metav1.GetOptions{})
		if getErr != nil {
			lastErr = fmt.Errorf("failed to get GlobalRole %s: %w", updatedGlobalRole.Name, getErr)
			return false, nil
		}

		updatedGlobalRole.ResourceVersion = current.ResourceVersion
		updated, lastErr = client.WranglerContext.Mgmt.GlobalRole().Update(updatedGlobalRole)
		if lastErr != nil {
			if errors.IsConflict(lastErr) {
				return false, nil
			}
			return false, lastErr
		}

		return true, nil
	},
	)

	if err != nil {
		return nil, fmt.Errorf("timed out updating GlobalRole %s: %w", updatedGlobalRole.Name, lastErr)
	}

	return updated, nil
}

// UpdateGlobalRoleBinding is a helper function that uses wrangler context to update an existing global role binding
func UpdateGlobalRoleBinding(client *rancher.Client, updatedGRB *v3.GlobalRoleBinding) (*v3.GlobalRoleBinding, error) {
	var updated *v3.GlobalRoleBinding
	var lastErr error

	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
		current, getErr := client.WranglerContext.Mgmt.GlobalRoleBinding().Get(updatedGRB.Name, metav1.GetOptions{})
		if getErr != nil {
			lastErr = fmt.Errorf("failed to get GlobalRoleBinding %s: %w", updatedGRB.Name, getErr)
			return false, nil
		}

		updatedGRB.ResourceVersion = current.ResourceVersion
		updated, lastErr = client.WranglerContext.Mgmt.GlobalRoleBinding().Update(updatedGRB)
		if lastErr != nil {
			if errors.IsConflict(lastErr) {
				return false, nil
			}
			return false, lastErr
		}

		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("timed out updating GlobalRoleBinding %s: %w", updatedGRB.Name, lastErr)
	}

	return updated, nil
}

// UpdateRoleTemplate is a helper function that uses wrangler context to update an existing role template
func UpdateRoleTemplate(client *rancher.Client, updatedRoleTemplate *v3.RoleTemplate) (*v3.RoleTemplate, error) {
	var updated *v3.RoleTemplate
	var lastErr error

	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
		current, getErr := client.WranglerContext.Mgmt.RoleTemplate().Get(updatedRoleTemplate.Name, metav1.GetOptions{})
		if getErr != nil {
			lastErr = fmt.Errorf("failed to get RoleTemplate %s: %w", updatedRoleTemplate.Name, getErr)
			return false, nil
		}

		updatedRoleTemplate.ResourceVersion = current.ResourceVersion
		updated, lastErr = client.WranglerContext.Mgmt.RoleTemplate().Update(updatedRoleTemplate)
		if lastErr != nil {
			if errors.IsConflict(lastErr) {
				return false, nil
			}
			return false, lastErr
		}

		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("timed out updating RoleTemplate %s: %w", updatedRoleTemplate.Name, lastErr)
	}

	return updated, nil
}

// UpdateClusterRoleTemplateBinding is a helper function that uses wrangler context to update an existing cluster role template binding
func UpdateClusterRoleTemplateBinding(client *rancher.Client, updatedCRTB *v3.ClusterRoleTemplateBinding) (*v3.ClusterRoleTemplateBinding, error) {
	var updated *v3.ClusterRoleTemplateBinding
	var lastErr error

	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
		current, getErr := client.WranglerContext.Mgmt.ClusterRoleTemplateBinding().Get(updatedCRTB.Namespace, updatedCRTB.Name, metav1.GetOptions{})
		if getErr != nil {
			lastErr = fmt.Errorf("failed to get ClusterRoleTemplateBinding %s/%s: %w", updatedCRTB.Namespace, updatedCRTB.Name, getErr)
			return false, nil
		}

		updatedCRTB.ResourceVersion = current.ResourceVersion
		updated, lastErr = client.WranglerContext.Mgmt.ClusterRoleTemplateBinding().Update(updatedCRTB)
		if lastErr != nil {
			if errors.IsConflict(lastErr) {
				return false, nil
			}
			return false, lastErr
		}

		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("timed out updating ClusterRoleTemplateBinding %s/%s: %w", updatedCRTB.Namespace, updatedCRTB.Name, lastErr)
	}

	return updated, nil
}

// UpdateProjectRoleTemplateBinding is a helper function that uses wrangler context to update an existing project role template binding
func UpdateProjectRoleTemplateBinding(client *rancher.Client, updatedPRTB *v3.ProjectRoleTemplateBinding) (*v3.ProjectRoleTemplateBinding, error) {
	var updated *v3.ProjectRoleTemplateBinding
	var lastErr error

	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
		current, getErr := client.WranglerContext.Mgmt.ProjectRoleTemplateBinding().Get(updatedPRTB.Namespace, updatedPRTB.Name, metav1.GetOptions{})
		if getErr != nil {
			lastErr = fmt.Errorf("failed to get ProjectRoleTemplateBinding %s/%s: %w", updatedPRTB.Namespace, updatedPRTB.Name, getErr)
			return false, nil
		}

		updatedPRTB.ResourceVersion = current.ResourceVersion
		updated, lastErr = client.WranglerContext.Mgmt.ProjectRoleTemplateBinding().Update(updatedPRTB)
		if lastErr != nil {
			if errors.IsConflict(lastErr) {
				return false, nil
			}
			return false, lastErr
		}

		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("timed out updating ProjectRoleTemplateBinding %s/%s: %w", updatedPRTB.Namespace, updatedPRTB.Name, lastErr)
	}

	return updated, nil
}
