package eks

import (
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
)

// CreateEKSHostedCluster is a helper function that creates an EKS hosted cluster
func CreateEKSHostedCluster(client *rancher.Client, displayName, cloudCredentialID string, eksClusterConfig ClusterConfig, enableClusterAlerting, enableClusterMonitoring, enableNetworkPolicy, windowsPreferedCluster bool, labels map[string]string) (*management.Cluster, error) {
	eksHostCluster := eksHostClusterConfig(displayName, cloudCredentialID, eksClusterConfig)
	cluster := &management.Cluster{
		DockerRootDir:          "/var/lib/docker",
		EKSConfig:              eksHostCluster,
		Name:                   displayName,
		EnableNetworkPolicy:    &enableNetworkPolicy,
		Labels:                 labels,
		WindowsPreferedCluster: windowsPreferedCluster,
	}

	clusterResp, err := client.Management.Cluster.Create(cluster)
	if err != nil {
		return nil, err
	}
	return clusterResp, err
}
