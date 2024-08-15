package cluster

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/wrangler/pkg/summary"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IsClusterActive is a helper function that uses the dynamic client to return cluster's ready state.
func IsClusterActive(client *rancher.Client, clusterID string) (ready bool, err error) {
	dynamic, err := client.GetRancherDynamicClient()
	if err != nil {
		return
	}

	unstructuredCluster, err := dynamic.Resource(ManagementClusterGVR()).Get(context.TODO(), clusterID, metav1.GetOptions{})
	if err != nil {
		return
	}

	summarized := summary.Summarize(unstructuredCluster)

	return summarized.IsReady(), nil
}
