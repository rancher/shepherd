package hpa

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	extclusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// DeleteHPA deletes a HorizontalPodAutoscaler by name in the given namespace using wrangler context. If waitForDeletion is true, it will wait until the HPA is deleted
func DeleteHPA(client *rancher.Client, clusterID, namespace, name string, waitForDeletion bool) error {
	clusterContext, err := extclusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return fmt.Errorf("failed to get cluster context: %w", err)
	}

	err = clusterContext.Autoscaling.HorizontalPodAutoscaler().Delete(namespace, name, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete HPA: %w", err)
	}

	if waitForDeletion {
		err = WaitForHPADeletion(client, clusterID, namespace, name)
		if err != nil {
			return fmt.Errorf("timed out waiting for HPA %s to be deleted: %w", name, err)
		}
	}

	return nil
}

// WaitForHPADeletion waits until the HPA is deleted and no longer retrievable.
func WaitForHPADeletion(client *rancher.Client, clusterID, hpaNamespace, hpaName string) error {
	return kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.FiveMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		_, pollErr := GetHPAByName(client, clusterID, hpaNamespace, hpaName)
		if pollErr != nil {
			if k8serrors.IsNotFound(pollErr) {
				return true, nil
			}
			return false, pollErr
		}
		return false, nil
	})
}
