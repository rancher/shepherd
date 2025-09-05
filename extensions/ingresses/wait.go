package ingresses

import (
	"context"
	"time"

	v1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/extensions/defaults/stevestates"
	"github.com/rancher/shepherd/extensions/defaults/stevetypes"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// WaitIngress waits for an Ingress to be ready in the downstream cluster
func WaitIngress(client *v1.Client, ingressResp *v1.SteveAPIObject, ingressName string) error {
	err := kwait.PollUntilContextTimeout(context.TODO(), 500*time.Millisecond, defaults.OneMinuteTimeout, true, func(ctx context.Context) (done bool, err error) {
		ingress, err := client.SteveType(stevetypes.Ingress).ByID(ingressResp.ID)
		if err != nil {
			return false, nil
		}

		if ingress.State.Name == stevestates.Active {
			return true, nil
		}

		return false, nil
	})

	return err
}
