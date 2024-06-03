package workloads

import (
	"context"
	"time"

	steveV1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

const (
	active           = "active"
	defaultNamespace = "default"
	port             = "port"
	ServiceType      = "service"
)

// CreateDeploymentWithService is a helper function to create a deployment and service in the downstream cluster.
func CreateDeploymentWithService(steveclient *steveV1.Client, wlName string, deployment *v1.Deployment, service corev1.Service) (*steveV1.SteveAPIObject, *steveV1.SteveAPIObject, error) {
	logrus.Infof("Creating deployment: %s", wlName)
	deploymentResp, err := steveclient.SteveType(DeploymentSteveType).Create(deployment)
	if err != nil {
		return nil, nil, err
	}

	err = kwait.PollUntilContextTimeout(context.TODO(), 500*time.Millisecond, defaults.OneMinuteTimeout, true, func(ctx context.Context) (done bool, err error) {
		deployment, err := steveclient.SteveType(DeploymentSteveType).ByID(deploymentResp.ID)
		if err != nil {
			return false, nil
		}

		if deployment.State.Name == active {
			logrus.Infof("Successfully created deployment: %s", wlName)

			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return nil, nil, err
	}

	logrus.Infof("Creating service: %s", service.Name)
	serviceResp, err := steveclient.SteveType(ServiceType).Create(service)
	if err != nil {
		return nil, nil, err
	}

	err = kwait.PollUntilContextTimeout(context.TODO(), 500*time.Millisecond, defaults.OneMinuteTimeout, true, func(ctx context.Context) (done bool, err error) {
		service, err := steveclient.SteveType(ServiceType).ByID(serviceResp.ID)
		if err != nil {
			return false, nil
		}

		if service.State.Name == active {
			logrus.Infof("Successfully created service: %s", service.Name)

			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return nil, nil, err
	}

	return deploymentResp, serviceResp, err
}
