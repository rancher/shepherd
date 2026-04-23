package cluster

import (
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/wrangler"
)

const (
	LocalCluster = "local"
)

// GetClusterWranglerContext is a helper function that returns the wrangler context for a specific cluster
func GetClusterWranglerContext(client *rancher.Client, clusterID string) (*wrangler.Context, error) {
	if clusterID == LocalCluster {
		return client.WranglerContext, nil
	}

	return client.WranglerContext.DownStreamClusterWranglerContext(clusterID)
}
