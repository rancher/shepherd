package gke

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

	var gkeConfig = clusterResp.GKEConfig
	if gkeConfig.NodePools == nil {
		return nil, fmt.Errorf("NodePools is empty")
	}

	nodePools := *gkeConfig.NodePools
	*nodePools[0].InitialNodeCount += *nodePool.InitialNodeCount

	gkeHostCluster := &management.Cluster{
		DockerRootDir:          "/var/lib/docker",
		EnableNetworkPolicy:    clusterResp.EnableNetworkPolicy,
		GKEConfig:              gkeConfig,
		Labels:                 clusterResp.Labels,
		Name:                   clusterResp.Name,
		WindowsPreferedCluster: clusterResp.WindowsPreferedCluster,
	}

	logrus.Infof("Scaling the node pool to %v total nodes", *nodePools[0].InitialNodeCount)
	updatedCluster, err := client.Management.Cluster.Update(clusterResp, gkeHostCluster)
	if err != nil {
		return nil, err
	}

	err = kwait.Poll(500*time.Millisecond, 10*time.Minute, func() (done bool, err error) {
		clusterResp, err := client.Management.Cluster.ByID(updatedCluster.ID)
		if err != nil {
			return false, err
		}

		if clusterResp.State == active && clusterResp.NodeCount == *nodePools[0].InitialNodeCount {
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return updatedCluster, nil
}

// ScalingGKENodePoolsNodes is a helper function that tests scaling of an EKS node pool by adding a new one and then deleting it.
func ScalingGKENodePoolsNodes(client *rancher.Client, cluster *management.Cluster, nodePool *NodePool) (*management.Cluster, error) {
	updatedCluster, err := updateNodePoolQuantity(client, cluster, nodePool)
	if err != nil {
		return nil, err
	}

	logrus.Infof("Node pool has been scaled!")

	return updatedCluster, nil
}
