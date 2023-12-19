package machinepools

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	apisV1 "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	"github.com/rancher/shepherd/clients/rancher"
	v1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/extensions/defaults"
	nodestat "github.com/rancher/shepherd/extensions/nodes"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

const (
	active = "active"
	pool   = "pool"
)

// MatchNodeRolesToMachinePool matches the role of machinePools to the nodeRoles.
func MatchNodeRolesToMachinePool(nodeRoles NodeRoles, machinePools []apisV1.RKEMachinePool) (int, int32) {
	count := int32(0)
	for index, machinePoolConfig := range machinePools {
		if nodeRoles.ControlPlane != machinePoolConfig.ControlPlaneRole {
			continue
		}
		if nodeRoles.Etcd != machinePoolConfig.EtcdRole {
			continue
		}
		if nodeRoles.Worker != machinePoolConfig.WorkerRole {
			continue
		}

		count += *machinePoolConfig.Quantity

		return index, count
	}

	return -1, count
}

// updateMachinePoolQuantity is a helper method that will update the desired machine pool with the latest quantity.
func updateMachinePoolQuantity(client *rancher.Client, cluster *v1.SteveAPIObject, nodeRoles NodeRoles) (*v1.SteveAPIObject, error) {
	updateCluster, err := client.Steve.SteveType("provisioning.cattle.io.cluster").ByID(cluster.ID)
	if err != nil {
		return nil, err
	}

	updatedCluster := new(apisV1.Cluster)
	err = v1.ConvertToK8sType(cluster, &updatedCluster)
	if err != nil {
		return nil, err
	}

	updatedCluster.ObjectMeta.ResourceVersion = updateCluster.ObjectMeta.ResourceVersion
	machineConfig, newQuantity := MatchNodeRolesToMachinePool(nodeRoles, updatedCluster.Spec.RKEConfig.MachinePools)

	newQuantity += nodeRoles.Quantity
	updatedCluster.Spec.RKEConfig.MachinePools[machineConfig].Quantity = &newQuantity

	logrus.Infof("Scaling the machine pool to %v total nodes", newQuantity)
	cluster, err = client.Steve.SteveType("provisioning.cattle.io.cluster").Update(cluster, updatedCluster)
	if err != nil {
		return nil, err
	}

	err = kwait.Poll(500*time.Millisecond, defaults.TenMinuteTimeout, func() (done bool, err error) {
		clusterResp, err := client.Steve.SteveType("provisioning.cattle.io.cluster").ByID(cluster.ID)
		if err != nil {
			return false, err
		}

		if clusterResp.ObjectMeta.State.Name == active && nodestat.AllManagementNodeReady(client, cluster.ID, defaults.ThirtyMinuteTimeout) == nil {
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return cluster, nil
}

// NewRKEMachinePool is a constructor that sets up a apisV1.RKEMachinePool object to be used to provision a cluster.
func NewRKEMachinePool(controlPlaneRole, etcdRole, workerRole, windowsRole bool, poolName string, quantity int32, machineConfig *v1.SteveAPIObject, hostnameLengthLimit int, drainBeforeDelete bool) apisV1.RKEMachinePool {
	machineConfigRef := &corev1.ObjectReference{
		Kind: machineConfig.Kind,
		Name: machineConfig.Name,
	}

	machinePool := apisV1.RKEMachinePool{
		ControlPlaneRole:  controlPlaneRole,
		EtcdRole:          etcdRole,
		WorkerRole:        workerRole,
		NodeConfig:        machineConfigRef,
		Name:              poolName,
		Quantity:          &quantity,
		DrainBeforeDelete: drainBeforeDelete,
	}

	if windowsRole {
		machinePool.Labels = map[string]string{
			"cattle.io/os": "windows",
		}
	}

	if hostnameLengthLimit > 0 {
		machinePool.HostnameLengthLimit = hostnameLengthLimit
	}

	return machinePool
}

type NodeRoles struct {
	ControlPlane      bool  `json:"controlplane,omitempty" yaml:"controlplane,omitempty"`
	Etcd              bool  `json:"etcd,omitempty" yaml:"etcd,omitempty"`
	Worker            bool  `json:"worker,omitempty" yaml:"worker,omitempty"`
	Windows           bool  `json:"windows,omitempty" yaml:"windows,omitempty"`
	Quantity          int32 `json:"quantity" yaml:"quantity"`
	DrainBeforeDelete bool  `json:"drainBeforeDelete" yaml:"drainBeforeDelete"`
}

type Roles struct {
	Roles []string `json:"roles,omitempty" yaml:"roles,omitempty"`
}

// HostnameTruncation is a struct that is used to set the hostname length limit for a cluster or its pools during provisioning
type HostnameTruncation struct {
	PoolNameLengthLimit    int
	ClusterNameLengthLimit int
	Name                   string
}

func (n NodeRoles) String() string {
	result := make([]string, 0, 3)
	if n.Quantity < 1 {
		return ""
	}
	if n.ControlPlane {
		result = append(result, "controlplane")
	}
	if n.Etcd {
		result = append(result, "etcd")
	}
	if n.Worker {
		result = append(result, "worker")
	}
	return fmt.Sprintf("%d %s", n.Quantity, strings.Join(result, "+"))
}

// CreateAllMachinePools is a helper method that will loop and setup multiple node pools with the defined machinePoolConfigs
func CreateAllMachinePools(nodeRoles []NodeRoles, rolesPerPool []string, machineConfigs []v1.SteveAPIObject, machinePoolConfigs []Roles, hostnameLengthLimits []HostnameTruncation) []apisV1.RKEMachinePool {
	machinePools := make([]apisV1.RKEMachinePool, 0, len(rolesPerPool))
	hostnameLengthLimit := 0

	for poolIndex, poolConfig := range nodeRoles {
		nodeRoleIndex := MatchRoleToPool(rolesPerPool[poolIndex], machinePoolConfigs)
		machineConfig := machineConfigs[nodeRoleIndex]

		poolName := pool + strconv.Itoa(poolIndex)
		if hostnameLengthLimits != nil && len(hostnameLengthLimits) >= poolIndex {
			hostnameLengthLimit = hostnameLengthLimits[poolIndex].PoolNameLengthLimit
			poolName = hostnameLengthLimits[poolIndex].Name
		}

		if !poolConfig.Windows {
			machinePool := NewRKEMachinePool(poolConfig.ControlPlane, poolConfig.Etcd, poolConfig.Worker, false, poolName, poolConfig.Quantity, &machineConfig, hostnameLengthLimit, poolConfig.DrainBeforeDelete)
			machinePools = append(machinePools, machinePool)
		} else {
			machinePool := NewRKEMachinePool(false, false, poolConfig.Windows, poolConfig.Windows, poolName, poolConfig.Quantity, &machineConfig, hostnameLengthLimit, poolConfig.DrainBeforeDelete)
			machinePools = append(machinePools, machinePool)
		}
	}
	return machinePools
}

// ScaleMachinePoolNodes is a helper method that will scale the machine pool to the desired quantity.
func ScaleMachinePoolNodes(client *rancher.Client, cluster *v1.SteveAPIObject, nodeRoles NodeRoles) (*v1.SteveAPIObject, error) {
	scaledClusterResp, err := updateMachinePoolQuantity(client, cluster, nodeRoles)
	if err != nil {
		return nil, err
	}

	logrus.Infof("Machine pool has been scaled!")

	return scaledClusterResp, nil
}

// MatchRoleToPool matches the role of a pool to the Roles of a machine. Returns the index of the matching Roles.
func MatchRoleToPool(poolRole string, allRoles []Roles) int {
	hasMatch := false
	for poolIndex, machineRole := range allRoles {
		for _, configRole := range machineRole.Roles {
			if strings.Contains(poolRole, configRole) {
				hasMatch = true
			}
		}
		if hasMatch {
			return poolIndex
		}
	}
	logrus.Warn("unable to match pool to role, likely missing [roles] in machineConfig")
	return -1
}
