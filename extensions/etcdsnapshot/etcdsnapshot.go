package etcdsnapshot

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	apisV1 "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	rkev1 "github.com/rancher/rancher/pkg/apis/rke.cattle.io/v1"
	"github.com/rancher/shepherd/clients/rancher"
	rancherv1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/sirupsen/logrus"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

const (
	ProvisioningSteveResouceType = "provisioning.cattle.io.cluster"
	SnapshotSteveResourceType    = "rke.cattle.io.etcdsnapshot"
	SnapshotClusterNameLabel     = "rke.cattle.io/cluster-name"
	fleetNamespace               = "fleet-default"
	localClusterName             = "local"
	active                       = "active"
	readyStatus                  = "Resource is ready"
)

// CreateRKE2K3SSnapshot is a helper function to create a snapshot on an RKE2 or k3s cluster.
// returns the list of snapshots and an error, if any.
func CreateRKE2K3SSnapshot(client *rancher.Client, clusterName string) ([]rancherv1.SteveAPIObject, error) {
	clusterObject, clusterSteveObject, err := clusters.GetProvisioningClusterByName(client, clusterName, fleetNamespace)
	if err != nil {
		return nil, err
	}

	if clusterObject.Spec.RKEConfig != nil {
		if clusterObject.Spec.RKEConfig.ETCDSnapshotCreate == nil {
			clusterObject.Spec.RKEConfig.ETCDSnapshotCreate = &rkev1.ETCDSnapshotCreate{
				Generation: 1,
			}
		} else {
			clusterObject.Spec.RKEConfig.ETCDSnapshotCreate = &rkev1.ETCDSnapshotCreate{
				Generation: clusterObject.Spec.RKEConfig.ETCDSnapshotCreate.Generation + 1,
			}
		}
	} else {
		clusterObject.Spec.RKEConfig = &apisV1.RKEConfig{
			ETCDSnapshotCreate: &rkev1.ETCDSnapshotCreate{
				Generation: 1,
			},
		}
	}

	logrus.Infof("Creating snapshot...")
	updatedCluster, err := client.Steve.SteveType(clusters.ProvisioningSteveResourceType).Update(clusterSteveObject, clusterObject)
	if err != nil {
		return nil, err
	}

	updateTimestamp := time.Now()
	err = clusters.WaitOnClusterAfterSnapshot(client, updatedCluster.ID)
	if err != nil {
		return nil, err
	}

	var snapshots []rancherv1.SteveAPIObject

	err = kwait.PollUntilContextTimeout(context.TODO(), 5*time.Second, defaults.FiveMinuteTimeout, true, func(ctx context.Context) (done bool, err error) {
		query, err := url.ParseQuery(fmt.Sprintf("labelSelector=%s=%s", SnapshotClusterNameLabel, clusterName))
		if err != nil {
			return false, nil
		}

		snapshotSteveObjList, err := client.Steve.SteveType(SnapshotSteveResourceType).List(query)
		if err != nil {
			return false, nil
		}

		if len(snapshotSteveObjList.Data) == 0 {
			return false, nil
		}

		snapshots = []rancherv1.SteveAPIObject{}
		for _, snapshot := range snapshotSteveObjList.Data {
			_, err = client.Steve.SteveType(SnapshotSteveResourceType).ByID(snapshot.ID)
			if err != nil {
				return false, nil
			}

			// snapshot time doesn't include nanoseconds, but time.Now() does. Rounding up by 1 Second.
			if snapshot.CreationTimestamp.Time.Add(time.Duration(time.Second)).Compare(updateTimestamp) > -1 {
				snapshots = append(snapshots, snapshot)
			}
		}

		if len(snapshots) == 0 {
			return false, nil
		}

		return true, nil
	})

	// not registering cleanup func; users do not delete snapshots through rancher

	return snapshots, err
}

// RestoreRKE2K3SSnapshot is a helper function to restore a snapshot on an RKE2 or k3s cluster. Returns error if any.
func RestoreRKE2K3SSnapshot(client *rancher.Client, snapshotRestore *rkev1.ETCDSnapshotRestore, clusterName string) error {
	_, existingSteveAPIObject, err := clusters.GetProvisioningClusterByName(client, clusterName, fleetNamespace)
	if err != nil {
		return err
	}
	steveWithUpdates := existingSteveAPIObject

	clusterSpec := &apisV1.ClusterSpec{}
	err = rancherv1.ConvertToK8sType(steveWithUpdates.Spec, clusterSpec)
	if err != nil {
		return err
	}

	clusterSpec.RKEConfig.ETCDSnapshotRestore = snapshotRestore

	steveWithUpdates.Spec = clusterSpec

	logrus.Infof("Restoring snapshot: %v", snapshotRestore.Name)
	updatedCluster, err := client.Steve.SteveType(ProvisioningSteveResouceType).Update(existingSteveAPIObject, steveWithUpdates)
	if err != nil {
		return err
	}

	err = kwait.PollUntilContextTimeout(context.TODO(), 1*time.Second, defaults.ThirtyMinuteTimeout, true, func(ctx context.Context) (done bool, err error) {
		clusterResp, err := client.Steve.SteveType(ProvisioningSteveResouceType).ByID(updatedCluster.ID)
		if err != nil {
			return false, err
		}

		if clusterResp.State.Error {
			return false, errors.New(clusterResp.State.Message)
		}

		clusterStatus := &apisV1.ClusterStatus{}
		err = rancherv1.ConvertToK8sType(clusterResp.Status, clusterStatus)
		if err != nil {
			return false, err
		}

		if clusterResp.State.Message == "waiting for etcd restore" {
			return true, nil
		}

		return false, nil
	})

	return err
}
