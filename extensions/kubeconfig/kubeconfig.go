package kubeconfig

import (
	"errors"

	"github.com/rancher/shepherd/clients/rancher"
	"k8s.io/client-go/tools/clientcmd"
)

// GetKubeconfig generates a kubeconfig froma specific cluster, and returns it in the form of a *clientcmd.ClientConfig
func GetKubeconfig(client *rancher.Client, clusterID string) (*clientcmd.ClientConfig, error) {
	cluster, err := client.Management.Cluster.ByID(clusterID)
	if err != nil {
		return nil, err
	}

	kubeConfig, err := client.Management.Cluster.ActionGenerateKubeconfig(cluster)
	if err != nil {
		return nil, err
	}

	configBytes := []byte(kubeConfig.Config)

	clientConfig, err := clientcmd.NewClientConfigFromBytes(configBytes)
	if err != nil {
		return nil, err
	}

	cfg, ok := clientConfig.(clientcmd.ClientConfig)
	if !ok {
		return nil, errors.New("error converting OverridingClientConfig to ClientConfig")
	}

	return &cfg, nil
}
