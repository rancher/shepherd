package steve

import (
	"context"
	"strings"
	"time"

	provv1 "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	"github.com/rancher/shepherd/clients/rancher"
	steveV1 "github.com/rancher/shepherd/clients/rancher/v1"
	v1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/extensions/defaults/stevetypes"
	"github.com/sirupsen/logrus"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

const (
	notFound     = "not found"
	localCluster = "local"
)

// WaitForResourceDeletion waits for a given steve object to be deleted.
func WaitForResourceDeletion(client *v1.Client, v1Resource *v1.SteveAPIObject, interval, timeout time.Duration) error {
	err := kwait.PollUntilContextTimeout(context.TODO(), interval, timeout, true, func(ctx context.Context) (done bool, err error) {
		_, err = client.SteveType(v1Resource.Type).ByID(v1Resource.ID)
		if err != nil {
			if strings.Contains(err.Error(), notFound) {
				logrus.Tracef("%s(%s) is deleted", v1Resource.Kind, v1Resource.Name)
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

// WaitForResourceState waits for a given steve object to be reach a desired state.
func WaitForResourceState(client *v1.Client, v1Resource *v1.SteveAPIObject, desiredState string, interval, timeout time.Duration) error {
	logrus.Tracef("Waiting for %s(%s) to reach a %s state", v1Resource.Kind, v1Resource.Name, desiredState)
	err := kwait.PollUntilContextTimeout(context.TODO(), interval, timeout, true, func(ctx context.Context) (done bool, err error) {
		clusterResp, err := client.SteveType(v1Resource.Type).ByID(v1Resource.ID)
		if err != nil {
			return false, err
		}

		if clusterResp.ObjectMeta.State.Name == desiredState {
			logrus.Tracef("%s(%s) is %s", v1Resource.Kind, v1Resource.Name, desiredState)
			return true, nil
		}

		return false, nil
	})

	return err
}

// CreateAndWaitForResource creates a steve resource and polls the resulting object until it is in the provided state.
func CreateAndWaitForResource(client *rancher.Client, clusterID, v1ResourceType string, v1Resource any, desiredState string, interval, timeout time.Duration) (*v1.SteveAPIObject, error) {
	steveClient, err := GetClusterClient(client, clusterID)
	if err != nil {
		return nil, err
	}

	resource, err := steveClient.SteveType(v1ResourceType).Create(v1Resource)
	if err != nil {
		return nil, err
	}

	logrus.Tracef("Creating %s(%s)", resource.Kind, resource.Name)

	if interval != 0 && timeout != 0 {
		err := WaitForResourceState(steveClient, resource, desiredState, interval, timeout)
		if err != nil {
			return resource, err
		}
	}

	return resource, nil
}

// GetClusterClient fetches the the client of a downstream cluster.
func GetClusterClient(client *rancher.Client, clusterID string) (*v1.Client, error) {
	var clusterClient *v1.Client
	var err error

	adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
	if err != nil {
		return nil, err
	}

	cluster, err := adminClient.Steve.SteveType(stevetypes.Provisioning).ByID(clusterID)
	if err != nil {
		return nil, err
	}

	status := &provv1.ClusterStatus{}
	err = steveV1.ConvertToK8sType(cluster.Status, status)
	if err != nil {
		return nil, err
	}

	clusterClient, err = adminClient.Steve.ProxyDownstream(status.ClusterName)
	if err != nil {
		return nil, err
	}

	return clusterClient, err
}
