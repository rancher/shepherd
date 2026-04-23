package ingresses

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	clusterapi "github.com/rancher/shepherd/extensions/kubeapi/cluster"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// GetIngress is a helper function that uses the wrangler context to get an ingress in a specific namespace for a specific cluster.
func GetIngress(client *rancher.Client, clusterID, namespace, ingressName string) (*networkingv1.Ingress, error) {
	wranglerCtx, err := clusterapi.GetClusterWranglerContext(client, clusterID)
	if err != nil {
		return nil, err
	}

	return wranglerCtx.Networking.Ingress().Get(namespace, ingressName, metav1.GetOptions{})
}

// IsIngressReady is a helper function that uses the wrangler context to determine if an ingress has a load balancer hostname or IP assigned.
func IsIngressReady(client *rancher.Client, clusterID, namespace, ingressName string) (bool, error) {
	ingressResp, err := GetIngress(client, clusterID, namespace, ingressName)
	if err != nil {
		return false, err
	}

	for _, ingressStatus := range ingressResp.Status.LoadBalancer.Ingress {
		if ingressStatus.Hostname != "" || ingressStatus.IP != "" {
			return true, nil
		}
	}

	return false, nil
}

// WaitForIngressReady is a helper function that uses the wrangler context to wait for an ingress to be ready in a specific namespace for a specific cluster.
func WaitForIngressReady(client *rancher.Client, clusterID, namespace, ingressName string) error {
	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveHundredMillisecondTimeout, defaults.OneMinuteTimeout, true, func(ctx context.Context) (done bool, err error) {
		isReady, err := IsIngressReady(client, clusterID, namespace, ingressName)
		if err != nil {
			return false, err
		}
		return isReady, nil
	})

	return err
}
