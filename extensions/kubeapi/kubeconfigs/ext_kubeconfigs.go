package kubeconfigs

import (
	"context"
	"fmt"

	extapi "github.com/rancher/rancher/pkg/apis/ext.cattle.io/v1"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// CreateKubeconfig creates a kubeconfig using wrangler context
func CreateKubeconfig(client *rancher.Client, kubeconfig *extapi.Kubeconfig) (*extapi.Kubeconfig, error) {
	createdKubeconfig, err := client.WranglerContext.Ext.Kubeconfig().Create(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubeconfig: %w", err)
	}

	return createdKubeconfig, nil
}

// GetKubeconfigByName retrieves a kubeconfig by name using the wrangler context
func GetKubeconfigByName(client *rancher.Client, kubeconfigName string) (*extapi.Kubeconfig, error) {
	kubeconfig, err := client.WranglerContext.Ext.Kubeconfig().Get(kubeconfigName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig %s: %w", kubeconfigName, err)
	}

	return kubeconfig, nil
}

// ListKubeconfig retrieves kubeconfig using the wrangler context and returns the list of kubeconfigs
func ListKubeconfigs(client *rancher.Client, listOpts metav1.ListOptions) (*extapi.KubeconfigList, error) {
	kubeconfigs, err := client.WranglerContext.Ext.Kubeconfig().List(listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list kubeconfig: %w", err)
	}

	return kubeconfigs, nil
}

// UpdateKubeconfig updates an existing kubeconfig using the wrangler context and returns the updated kubeconfig object
func UpdateKubeconfig(client *rancher.Client, kubeconfig *extapi.Kubeconfig) (*extapi.Kubeconfig, error) {
	if kubeconfig == nil {
		return nil, fmt.Errorf("kubeconfig object is nil")
	}

	var updated *extapi.Kubeconfig
	var lastErr error

	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		current, getErr := GetKubeconfigByName(client, kubeconfig.Name)
		if getErr != nil {
			lastErr = fmt.Errorf("failed to get kubeconfig %s: %w", kubeconfig.Name, getErr)
			return false, nil
		}
		kubeconfig.ResourceVersion = current.ResourceVersion
		updated, lastErr = client.WranglerContext.Ext.Kubeconfig().Update(kubeconfig)
		if lastErr != nil {
			if k8serrors.IsConflict(lastErr) {
				return false, nil
			}
			return false, lastErr
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("timed out updating kubeconfig %s: %w", kubeconfig.Name, lastErr)
	}

	return updated, nil
}

// DeleteKubeconfig deletes a kubeconfig by name using the wrangler context and waits for the deletion to complete
func DeleteKubeconfig(client *rancher.Client, kubeconfigName string, waitForDelete bool) error {
	err := client.WranglerContext.Ext.Kubeconfig().Delete(kubeconfigName, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete kubeconfig %s: %w", kubeconfigName, err)
	}

	if waitForDelete {
		err = WaitForKubeconfigDeletion(client, kubeconfigName)
		if err != nil {
			return fmt.Errorf("timed out waiting for kubeconfig %s to be deleted: %w", kubeconfigName, err)
		}
	}

	return nil
}

// WaitForKubeconfigDeletion polls until the kubeconfig with the given name is deleted or the timeout is reached.
func WaitForKubeconfigDeletion(client *rancher.Client, kubeconfigName string) error {
	return kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		_, err := GetKubeconfigByName(client, kubeconfigName)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	})
}
