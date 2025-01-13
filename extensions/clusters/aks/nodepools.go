package aks

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	kwait "k8s.io/apimachinery/pkg/util/wait"

	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
)

const (
	active = "active"
)

// updateNodePoolQuantity is a helper method that will update the node pool with the desired quantity.
func updateNodePoolQuantity(client *rancher.Client, cluster *management.Cluster, nodePool *NodePool) (*management.Cluster, error) {
	clusterResp, err := client.Management.Cluster.ByID(cluster.ID)
	if err != nil {
		return nil, err
	}

	var aksConfig = clusterResp.AKSConfig

	if aksConfig.NodePools == nil {
		return nil, fmt.Errorf("NodePools is empty")
	}

	nodePools := *aksConfig.NodePools
	*nodePools[0].Count += *nodePool.NodeCount

	aksHostCluster := &management.Cluster{
		AKSConfig:              aksConfig,
		DockerRootDir:          "/var/lib/docker",
		EnableNetworkPolicy:    clusterResp.EnableNetworkPolicy,
		Labels:                 clusterResp.Labels,
		Name:                   clusterResp.Name,
		WindowsPreferedCluster: clusterResp.WindowsPreferedCluster,
	}

	logrus.Infof("Scaling the agentpool to %v total nodes", *nodePools[0].Count)
	updatedCluster, err := client.Management.Cluster.Update(clusterResp, aksHostCluster)
	if err != nil {
		return nil, err
	}

	err = kwait.Poll(500*time.Millisecond, 10*time.Minute, func() (done bool, err error) {
		clusterResp, err := client.Management.Cluster.ByID(updatedCluster.ID)
		if err != nil {
			return false, err
		}

		if clusterResp.State == active && clusterResp.NodeCount == *nodePools[0].Count {
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return updatedCluster, nil
}

// ScalingAKSNodePoolsNodes is a helper function that tests scaling of an AKS node pool by adding a new one and then deleting it.
func ScalingAKSNodePoolsNodes(client *rancher.Client, cluster *management.Cluster, nodePool *NodePool) (*management.Cluster, error) {
	updatedCluster, err := updateNodePoolQuantity(client, cluster, nodePool)
	if err != nil {
		return nil, err
	}

	logrus.Infof("Agentpool has been scaled!")

	return updatedCluster, nil
}
