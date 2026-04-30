package namespaces

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	clusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// UpdateNamespace is a helper function that uses wrangler context to update an existing namespace in a cluster
func UpdateNamespace(client *rancher.Client, clusterID string, updatedNamespace *corev1.Namespace) (*corev1.Namespace, error) {
	wranglerCtx, err := clusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	var updated *corev1.Namespace
	var lastErr error
	err = kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
		current, getErr := wranglerCtx.Core.Namespace().Get(updatedNamespace.Name, metav1.GetOptions{})
		if getErr != nil {
			lastErr = fmt.Errorf("failed to get Namespace %s: %w", updatedNamespace.Name, getErr)
			return false, nil
		}

		updatedNamespace.ResourceVersion = current.ResourceVersion
		updated, lastErr = wranglerCtx.Core.Namespace().Update(updatedNamespace)
		if lastErr != nil {
			if errors.IsConflict(lastErr) {
				return false, nil
			}
			return false, lastErr
		}

		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("timed out updating Namespace %s: %w", updatedNamespace.Name, lastErr)
	}

	return updated, nil
}
