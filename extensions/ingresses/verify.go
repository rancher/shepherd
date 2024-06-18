package ingresses

import (
	"context"
	"time"

	steveV1 "github.com/rancher/shepherd/clients/rancher/v1"
	v1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/sirupsen/logrus"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// VerifyIngress waits for an Ingress to be ready in the downstream cluster
func VerifyIngress(client *steveV1.Client, ingressResp *v1.SteveAPIObject, ingressName string) error {
	err := kwait.PollUntilContextTimeout(context.TODO(), 500*time.Millisecond, defaults.OneMinuteTimeout, true, func(ctx context.Context) (done bool, err error) {
		ingress, err := client.SteveType(IngressSteveType).ByID(ingressResp.ID)
		if err != nil {
			return false, nil
		}

		if ingress.State.Name == active {
			logrus.Infof("Successfully created ingress: %v", ingressName)

			return true, nil
		}

		return false, nil
	})

	return err
}
