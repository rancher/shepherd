package etcdsnapshot

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rancher/norman/types"
	apisV1 "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	rkev1 "github.com/rancher/rancher/pkg/apis/rke.cattle.io/v1"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	rancherv1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/extensions/defaults/stevetypes"
	"github.com/rancher/shepherd/extensions/kubeapi/nodes"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	ProvisioningSteveResouceType = "provisioning.cattle.io.cluster"
	fleetNamespace               = "fleet-default"
	localClusterName             = "local"
	active                       = "active"
)

func MatchNodeToAnyEtcdRole(client *rancher.Client, clusterID string) (int, *management.Node) {
	machines, err := client.Management.Node.List(&types.ListOpts{Filters: map[string]interface{}{
		"clusterId": clusterID,
	}})
	if err != nil {
		return 0, nil
	}

	numOfNodes := 0
	lastMatchingNode := &management.Node{}

	for _, machine := range machines.Data {
		if machine.Etcd {
			lastMatchingNode = &machine
			numOfNodes++
		}
	}

	return numOfNodes, lastMatchingNode
}

// GetRKE1Snapshots is a helper function to get the existing snapshots for a downstream RKE1 cluster.
func GetRKE1Snapshots(client *rancher.Client, clusterName string) ([]management.EtcdBackup, error) {
	clusterID, err := clusters.GetClusterIDByName(client, clusterName)
	if err != nil {
		return nil, err
	}

	snapshotSteveObjList, err := client.Management.EtcdBackup.ListAll(&types.ListOpts{
		Filters: map[string]interface{}{
			"clusterId": clusterID,
		},
	})
	if err != nil {
		return nil, err
	}

	snapshots := []management.EtcdBackup{}

	for _, snapshot := range snapshotSteveObjList.Data {
		if strings.Contains(snapshot.Name, clusterID) {
			snapshots = append(snapshots, snapshot)
		}
	}

	return snapshots, nil
}

// GetRKE2K3SSnapshots is a helper function to get the existing snapshots for a downstream RKE2/K3S cluster.
func GetRKE2K3SSnapshots(client *rancher.Client, localclusterID string, clusterName string) ([]rancherv1.SteveAPIObject, error) {
	steveclient, err := client.Steve.ProxyDownstream(localclusterID)
	if err != nil {
		return nil, err
	}

	snapshotSteveObjList, err := steveclient.SteveType(stevetypes.EtcdSnapshot).List(nil)
	if err != nil {
		return nil, err
	}

	snapshots := []rancherv1.SteveAPIObject{}

	for _, snapshot := range snapshotSteveObjList.Data {
		if strings.Contains(snapshot.ObjectMeta.Name, clusterName) {
			snapshots = append(snapshots, snapshot)
		}
	}

	return snapshots, nil
}

// CreateRKE1Snapshot is a helper function to create a snapshot on an RKE1 cluster. Returns error if any.
func CreateRKE1Snapshot(client *rancher.Client, clusterName string) error {
	clusterID, err := clusters.GetClusterIDByName(client, clusterName)
	if err != nil {
		return err
	}

	clusterResp, err := client.Management.Cluster.ByID(clusterID)
	if err != nil {
		return err
	}

	logrus.Infof("Creating snapshot...")
	err = client.Management.Cluster.ActionBackupEtcd(clusterResp)
	if err != nil {
		return err
	}

	err = wait.Poll(1*time.Second, defaults.FiveMinuteTimeout, func() (bool, error) {
		snapshotSteveObjList, err := client.Management.EtcdBackup.ListAll(&types.ListOpts{
			Filters: map[string]interface{}{
				"clusterId": clusterID,
			},
		})
		if err != nil {
			return false, nil
		}

		for _, snapshot := range snapshotSteveObjList.Data {
			snapshotObj, err := client.Management.EtcdBackup.ByID(snapshot.ID)
			if err != nil {
				return false, nil
			}

			if snapshotObj.State != active {
				return false, nil
			}
		}

		logrus.Infof("All snapshots in the cluster are in an active state!")
		return true, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// CreateRKE2K3SSnapshot is a helper function to create a snapshot on an RKE2 or k3s cluster. Returns error if any.
func CreateRKE2K3SSnapshot(client *rancher.Client, clusterName string) error {
	clusterObject, clusterSteveObject, err := clusters.GetProvisioningClusterByName(client, clusterName, fleetNamespace)
	if err != nil {
		return err
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
	_, err = client.Steve.SteveType(clusters.ProvisioningSteveResourceType).Update(clusterSteveObject, clusterObject)
	if err != nil {
		return err
	}

	err = wait.Poll(1*time.Second, defaults.FiveMinuteTimeout, func() (bool, error) {
		snapshotSteveObjList, err := client.Steve.SteveType("rke.cattle.io.etcdsnapshot").List(nil)
		if err != nil {
			return false, nil
		}

		_, clusterSteveObject, err := clusters.GetProvisioningClusterByName(client, clusterName, fleetNamespace)
		if err != nil {
			return false, nil
		}

		for _, snapshot := range snapshotSteveObjList.Data {
			snapshotObj, err := client.Steve.SteveType("rke.cattle.io.etcdsnapshot").ByID(snapshot.ID)
			if err != nil {
				return false, nil
			}

			if snapshotObj.ObjectMeta.State.Name == active && clusterSteveObject.ObjectMeta.State.Name == active {
				logrus.Infof("All snapshots in the cluster are in an active state!")
				return true, nil
			}
		}

		return false, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// RestoreRKE1Snapshot is a helper function to restore a snapshot on an RKE1 cluster. Returns error if any.
func RestoreRKE1Snapshot(client *rancher.Client, clusterName string, snapshotRestore *management.RestoreFromEtcdBackupInput, initialControlPlaneValue, initialWorkerValue string) error {
	clusterID, err := clusters.GetClusterIDByName(client, clusterName)
	if err != nil {
		return err
	}

	cluster, err := client.Management.Cluster.ByID(clusterID)
	if err != nil {
		return err
	}

	updatedCluster := cluster

	updatedCluster.RancherKubernetesEngineConfig.UpgradeStrategy.MaxUnavailableControlplane = initialControlPlaneValue
	updatedCluster.RancherKubernetesEngineConfig.UpgradeStrategy.MaxUnavailableWorker = initialWorkerValue

	_, err = client.Management.Cluster.Update(cluster, updatedCluster)
	if err != nil {
		return err
	}

	logrus.Infof("Restoring snapshot: %v", snapshotRestore.EtcdBackupID)
	err = client.Management.Cluster.ActionRestoreFromEtcdBackup(cluster, snapshotRestore)
	if err != nil {
		return err
	}

	err = wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, defaults.ThirtyMinuteTimeout, true, func(ctx context.Context) (done bool, err error) {
		clusterResp, err := client.Management.Cluster.ByID(clusterID)
		if err != nil {
			return false, nil
		}

		if clusterResp.State == active {
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// RestoreRKE2K3SSnapshot is a helper function to restore a snapshot on an RKE2 or k3s cluster. Returns error if any.
func RestoreRKE2K3SSnapshot(client *rancher.Client, clusterName string, snapshotRestore *rkev1.ETCDSnapshotRestore, initialControlPlaneValue, initialWorkerValue string) error {
	clusterObject, existingSteveAPIObject, err := clusters.GetProvisioningClusterByName(client, clusterName, fleetNamespace)
	if err != nil {
		return err
	}

	clusterObject.Spec.RKEConfig.ETCDSnapshotRestore = snapshotRestore
	clusterObject.Spec.RKEConfig.UpgradeStrategy.ControlPlaneConcurrency = initialControlPlaneValue
	clusterObject.Spec.RKEConfig.UpgradeStrategy.WorkerConcurrency = initialWorkerValue

	logrus.Infof("Restoring snapshot: %v", snapshotRestore.Name)
	_, err = client.Steve.SteveType(ProvisioningSteveResouceType).Update(existingSteveAPIObject, clusterObject)
	if err != nil {
		return err
	}

	return nil
}

// RKE1RetentionLimitCheck is a check that validates that the number of automatic snapshots on the cluster is under the retention limit
func RKE1RetentionLimitCheck(client *rancher.Client, clusterName string) error {
	clusterID, err := clusters.GetClusterIDByName(client, clusterName)
	if err != nil {
		return err
	}

	clusterResp, err := client.Management.Cluster.ByID(clusterID)
	if err != nil {
		return err
	}

	retentionLimit := clusterResp.RancherKubernetesEngineConfig.Services.Etcd.BackupConfig.Retention
	s3Config := clusterResp.RancherKubernetesEngineConfig.Services.Etcd.BackupConfig.S3BackupConfig

	isS3 := false
	if s3Config != nil {
		isS3 = true
	}

	existingSnapshots, err := GetRKE1Snapshots(client, clusterName)
	if err != nil {
		return err
	}

	automaticSnapshots := []management.EtcdBackup{}

	for _, snapshot := range existingSnapshots {
		if !snapshot.Manual {
			automaticSnapshots = append(automaticSnapshots, snapshot)
		}
	}

	listOpts := metav1.ListOptions{LabelSelector: "node-role.kubernetes.io/etcd=true"}
	etcdNodes, err := nodes.GetNodes(client, clusterID, listOpts)
	if err != nil {
		return err
	}

	expectedSnapshotsNum := int(retentionLimit) * len(etcdNodes)
	if isS3 {
		expectedSnapshotsNum = expectedSnapshotsNum * 2
	}

	if len(automaticSnapshots) > expectedSnapshotsNum {
		errMsg := fmt.Sprintf("retention limit exceeded: expected %d snapshots, found %d snapshots",
			expectedSnapshotsNum, len(automaticSnapshots))

		return errors.New(errMsg)
	}

	logrus.Infof("Snapshot retention limit respected, Snapshots Expected: %v Snapshots Found: %v",
		expectedSnapshotsNum, len(automaticSnapshots))

	return nil
}

// RKE2K3SRetentionLimitCheck is a check that validates that the number of automatic snapshots
// on the cluster is under the retention limit.
func RKE2K3SRetentionLimitCheck(client *rancher.Client, clusterName string) error {
	v1ClusterID, err := clusters.GetV1ProvisioningClusterByName(client, clusterName)
	if err != nil {
		return err
	}

	clusterObj, err := client.Steve.SteveType(stevetypes.Provisioning).ByID(v1ClusterID)
	if err != nil {
		return err
	}

	spec := apisV1.ClusterSpec{}
	err = rancherv1.ConvertToK8sType(clusterObj.Spec, &spec)
	if err != nil {
		return err
	}

	etcdConfig := spec.RKEConfig.ETCD
	retentionLimit := etcdConfig.SnapshotRetention

	isS3 := false
	if etcdConfig.S3 != nil {
		isS3 = true
	}

	localClusterID, err := clusters.GetClusterIDByName(client, localClusterName)
	if err != nil {
		return err
	}

	existingSnapshots, err := GetRKE2K3SSnapshots(client, localClusterID, clusterName)
	if err != nil {
		return err
	}

	automaticSnapshots := []rancherv1.SteveAPIObject{}

	for _, snapshot := range existingSnapshots {
		if strings.Contains(snapshot.Annotations["etcdsnapshot.rke.io/snapshot-file-name"], "etcd-snapshot") {
			automaticSnapshots = append(automaticSnapshots, snapshot)
		}
	}

	downstreamClusterID, err := clusters.GetClusterIDByName(client, clusterName)
	if err != nil {
		return err
	}

	listOpts := metav1.ListOptions{LabelSelector: "node-role.kubernetes.io/etcd=true"}
	etcdNodes, err := nodes.GetNodes(client, downstreamClusterID, listOpts)
	if err != nil {
		return err
	}

	expectedSnapshotsNum := int(retentionLimit) * len(etcdNodes)
	if isS3 {
		expectedSnapshotsNum = expectedSnapshotsNum * 2
	}

	if len(automaticSnapshots) > expectedSnapshotsNum {
		msg := fmt.Sprintf(
			"retention limit exceeded: expected %d snapshots, found %d snapshots",
			expectedSnapshotsNum, len(automaticSnapshots))

		return errors.New(msg)
	}

	logrus.Infof("Snapshot retention limit respected, Snapshots Expected: %v Snapshots Found: %v",
		expectedSnapshotsNum, len(automaticSnapshots))

	return nil
}
