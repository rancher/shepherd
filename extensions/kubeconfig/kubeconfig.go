package kubeconfig

import (
	"errors"
	"os"

	"github.com/pkg/errors"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
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

func GetKubeconfigFromFlags(masterURL, kubeconfigPath string) (*clientcmd.ClientConfig, error) {
	if _, err := os.Stat(kubeconfigPath); err != nil {
		return nil, errors.Wrap(err, "GetKubeconfigFromFlags: ")
	}
	kubeConfigContent, err := os.ReadFile(kubeconfigPath) //read the content of file
	if err != nil {
		return nil, err
	}

	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeConfigContent)
	if err != nil {
		return nil, err
	}
	if masterURL != "" {
		rawConfig, err := clientConfig.RawConfig()
		if err != nil {
			return nil, err
		}

		clientConfig = clientcmd.NewDefaultClientConfig(rawConfig, &clientcmd.ConfigOverrides{
			ClusterInfo: api.Cluster{
				Server: masterURL,
			},
		})
	}
	return &clientConfig, err
}

func GenerateKubeconfigForRestConfig(restConfig *rest.Config, defaultUser, defaultContext, clusterName string) ([]byte, error) {
	if defaultUser == "" || defaultContext == "" || clusterName == "" {
		return nil, errors.New("GenerateKubeconfigForRestConfig: 'defaultUser', 'defaultContext', and 'clusterName' must all be non-zero strings")
	}
	clusters := make(map[string]*api.Cluster)
	clusters["default-cluster"] = &api.Cluster{
		Server:                   restConfig.Host,
		CertificateAuthorityData: restConfig.CAData,
	}
	contexts := make(map[string]*api.Context)
	contexts["default-context"] = &api.Context{
		Cluster:  clusterName,
		AuthInfo: defaultUser,
	}
	authinfos := make(map[string]*api.AuthInfo)
	authinfos["default-user"] = &api.AuthInfo{
		ClientCertificateData: restConfig.CertData,
		ClientKeyData:         restConfig.KeyData,
	}
	clientConfig := api.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: defaultContext,
		AuthInfos:      authinfos,
	}
	return clientcmd.Write(clientConfig)
}

func GetKubeConfigBytes(client *rancher.Client, clusterID string) ([]byte, error) {
	cluster, err := client.Management.Cluster.ByID(clusterID)
	if err != nil {
		return nil, err
	}

	kubeConfig, err := client.Management.Cluster.ActionGenerateKubeconfig(cluster)
	if err != nil {
		return nil, err
	}

	return []byte(kubeConfig.Config), err
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
