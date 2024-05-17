package steve

import (
	"context"
	"strings"
	"time"

	"github.com/rancher/shepherd/clients/rancher"
	v1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/sirupsen/logrus"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

const (
	notFound = "not found"
	active   = "active"
)

// WaitForSteveResourceDeletion accepts a client, steve resource type, and resource ID, then waits for a steve resource to be deleted
func WaitForSteveResourceDeletion(client *rancher.Client, interval, timeout time.Duration, steveResourceType, steveID string) error {
	err := kwait.PollUntilContextTimeout(context.TODO(), interval, timeout, true, func(ctx context.Context) (done bool, err error) {
		_, err = client.Steve.SteveType(steveResourceType).ByID(steveID)
		if err != nil {
			if strings.Contains(err.Error(), notFound) {
				logrus.Info("Resource was successfully removed!")
				return true, nil
			} else {
				return false, err
			}
		}

		return false, nil
	})

	if err != nil {
		return err
	}

	return nil
}

// WaitForSteveResourceCreation waits for a given steve object to be created/come up active.
func WaitForSteveResourceCreation(client *rancher.Client, v1Resource *v1.SteveAPIObject, interval, timeout time.Duration) error {
	err := kwait.PollUntilContextTimeout(context.TODO(), interval, timeout, true, func(ctx context.Context) (done bool, err error) {
		client, err = client.ReLogin()
		if err != nil {
			return false, err
		}

		clusterResp, err := client.Steve.SteveType(v1Resource.Type).ByID(v1Resource.ID)
		if err != nil {
			return false, err
		}

		if clusterResp.ObjectMeta.State.Name == active {
			logrus.Infof("%s(%s) is active", v1Resource.Kind, v1Resource.Name)
			return true, nil
		}

		return false, nil
	})

	return err
}

// CreateResource creates a steve resource and polls the resulting object until it comes up active.
func CreateAndWaitForResource(client *rancher.Client, v1ResourceType string, v1Resource any, poll bool, interval, timeout time.Duration) (*v1.SteveAPIObject, error) {
	resource, err := client.Steve.SteveType(v1ResourceType).Create(v1Resource)
	if err != nil {
		return nil, err
	}
	logrus.Infof("Creating %s(%s)", resource.Kind, resource.Name)

	if poll {
		err := WaitForSteveResourceCreation(client, resource, interval, timeout)
		if err != nil {
			return resource, err
		}
	}

	return resource, nil
}
