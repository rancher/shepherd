package secrets

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// UpdateSecret updates a Secret in the given namespace using wrangler context, handling conflicts with polling.
func UpdateSecret(client *rancher.Client, clusterID string, secret *corev1.Secret) (*corev1.Secret, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	var updated *corev1.Secret
	var lastErr error

	err = kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		current, getErr := GetSecretByName(client, clusterID, secret.Namespace, secret.Name)
		if getErr != nil {
			lastErr = fmt.Errorf("failed to get Secret %s/%s: %w", secret.Namespace, secret.Name, getErr)
			return false, nil
		}
		secret.ResourceVersion = current.ResourceVersion
		updated, lastErr = clusterContext.Core.Secret().Update(secret)
		if lastErr != nil {
			if errors.IsConflict(lastErr) {
				return false, nil
			}
			return false, lastErr
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("timed out updating Secret %s/%s: %w", secret.Namespace, secret.Name, lastErr)
	}

	return updated, nil
}
