package configmaps

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

// UpdateConfigMap updates an existing ConfigMap in the specified namespace and cluster using wrangler context, handling conflicts with polling.
func UpdateConfigMap(client *rancher.Client, clusterID, namespace string, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster context: %w", err)
	}

	var updated *corev1.ConfigMap
	var lastErr error

	err = kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		current, getErr := GetConfigMapByName(client, clusterID, namespace, configMap.Name)
		if getErr != nil {
			lastErr = fmt.Errorf("failed to get ConfigMap %s/%s: %w", namespace, configMap.Name, getErr)
			return false, nil
		}
		configMap.ResourceVersion = current.ResourceVersion
		updated, lastErr = clusterContext.Core.ConfigMap().Update(configMap)
		if lastErr != nil {
			if errors.IsConflict(lastErr) {
				return false, nil
			}
			return false, lastErr
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("timed out updating ConfigMap %s/%s: %w", namespace, configMap.Name, lastErr)
	}

	return updated, nil
}
