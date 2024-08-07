package kubeconfig

import (
	"errors"
	"os"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/sirupsen/logrus"
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

func WriteKubeconfigToFile(client *rancher.Client, clusterID, path string) error {
	cluster, err := client.Management.Cluster.ByID(clusterID)
	if err != nil {
		return err
	}

	kubeConfig, err := client.Management.Cluster.ActionGenerateKubeconfig(cluster)
	if err != nil {
		return err
	}

	_, statErr := os.Stat(path)
	if statErr == nil {
		err = os.Remove(path)
	}

	if f, err := os.Create(path); err == nil {
		if _, err := f.Write([]byte(kubeConfig.Config)); err != nil {
			return err
		}
	}
	logrus.Infof("Finished writing. Err: %v", err)
	return err
}
