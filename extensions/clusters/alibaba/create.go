package alibaba

import (
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
)

// CreateAlibabaHostedCluster is a helper function that creates an Alibaba hosted cluster
func CreateAlibabaHostedCluster(client *rancher.Client, displayName, cloudCredentialID string, aliClusterConfig ClusterConfig, enableClusterAlerting, enableClusterMonitoring, enableNetworkPolicy, windowsPreferedCluster bool, labels map[string]string) (*management.Cluster, error) {
	aliHostCluster := aliHostClusterConfig(displayName, cloudCredentialID, aliClusterConfig)
	cluster := &management.Cluster{
		DockerRootDir:          "/var/lib/docker",
		AliConfig:              aliHostCluster,
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
