package configmaps

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	"k8s.io/apimachinery/pkg/api/errors"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// DeleteConfigMap deletes a ConfigMap by name in the specified namespace and cluster using wrangler context.
func DeleteConfigMap(client *rancher.Client, clusterID, namespace, name string, waitForDelete bool) error {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return err
	}

	err = clusterContext.Core.ConfigMap().Delete(namespace, name, nil)
	if err != nil {
		return err
	}

	if waitForDelete {
		return WaitForConfigMapDeletion(client, clusterID, namespace, name)
	}

	return nil
}

// WaitForConfigMapDeletion waits for a ConfigMap to be deleted from the specified namespace using the wrangler context for the given cluster
func WaitForConfigMapDeletion(client *rancher.Client, clusterID, namespace, configMapName string) error {
	return kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
		_, err = GetConfigMapByName(client, clusterID, namespace, configMapName)
		if errors.IsNotFound(err) {
			return true, nil
		}

		return false, err
	})
}
