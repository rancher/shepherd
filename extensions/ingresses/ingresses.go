package ingresses

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rancher/shepherd/clients/rancher"
	v1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/extensions/defaults/stevetypes"
	"github.com/rancher/shepherd/extensions/workloads/pods"
	"github.com/sirupsen/logrus"
	networking "k8s.io/api/networking/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

const (
	IngressSteveType = "networking.k8s.io.ingress"
	pod              = "pod"
	IngressNginx     = "ingress-nginx"
	RancherWebhook   = "rancher-webhook"
)

// GetExternalIngressResponse gets a response from a specific hostname and path.
// Returns the response and an error if any.
func GetExternalIngressResponse(client *rancher.Client, hostname string, path string, isWithTLS bool) (body string, err error) {
	protocol := "http"

	if isWithTLS {
		protocol = "https"
	}

	url := fmt.Sprintf("%s://%s/%s", protocol, hostname, path)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	req.Header.Add("Authorization", "Bearer "+client.RancherConfig.AdminToken)

	resp, err := client.Management.Ops.Client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		body = string(bodyBytes)
	} else {
		return "", errors.Wrapf(err, "resp status code is: %v while getting external ingress response", resp.StatusCode)
	}

	return
}

// IsIngressExternallyAccessible checks if the ingress is accessible externally,
// it returns true if the ingress is accessible, false if it is not, and an error if there is an error.
func IsIngressExternallyAccessible(client *rancher.Client, hostname string, path string, isWithTLS bool) (accessible bool, err error) {
	_, err = GetExternalIngressResponse(client, hostname, path, isWithTLS)
	if err != nil {
		return
	}

	return !accessible, nil
}

// CreateIngress will create an Ingress object in the downstream cluster.
func CreateIngress(client *v1.Client, ingressName string, ingressTemplate networking.Ingress) (*v1.SteveAPIObject, error) {
	podClient := client.SteveType(stevetypes.Pod)
	err := kwait.PollUntilContextTimeout(context.TODO(), 15*time.Second, defaults.FiveMinuteTimeout, true, func(context.Context) (done bool, err error) {
		newPods, err := podClient.List(nil)
		if err != nil {
			return false, nil
		}
		if len(newPods.Data) != 0 {
			return true, nil
		}
		for _, pod := range newPods.Data {
			if strings.Contains(pod.Name, IngressNginx) || strings.Contains(pod.Name, RancherWebhook) {
				isReady, podError := pods.IsPodReady(&pod)

				if podError != nil {
					return false, nil
				}

				return isReady, nil
			}
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	logrus.Infof("Create Ingress: %v", ingressName)
	ingressResp, err := client.SteveType(stevetypes.Ingress).Create(ingressTemplate)
	if err != nil {
		logrus.Errorf("Failed to create ingress: %v", err)

		return nil, err
	}

	return ingressResp, err
}
