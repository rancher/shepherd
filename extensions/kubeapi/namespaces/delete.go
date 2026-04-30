package namespaces

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	clusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// DeleteNamespace deletes a namespace in a cluster using wrangler context
func DeleteNamespace(client *rancher.Client, clusterID, namespaceName string) error {
	ctx, err := clusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return err
	}

	err = ctx.Core.Namespace().Delete(namespaceName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(context.Context) (bool, error) {
		_, pollErr := ctx.Core.Namespace().Get(namespaceName, metav1.GetOptions{})
		if pollErr != nil {
			if k8serrors.IsNotFound(pollErr) {
				return true, nil
			}
			return false, pollErr
		}
		return false, nil
	},
	)
}
