package cluster

import (
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/wrangler"
)

const (
	LocalCluster = "local"
)

// GetClusterWranglerContext returns the context for the cluster
func GetClusterWranglerContext(client *rancher.Client, clusterID string) (*wrangler.Context, error) {
	if clusterID == LocalCluster {
		return client.WranglerContext, nil
	}

	ctx, err := client.WranglerContext.DownStreamClusterWranglerContext(clusterID)
	if err != nil {
		return nil, err
	}

	return ctx, nil
}