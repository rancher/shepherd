package kubeconfig

import (
	"os"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

type RestGetter struct {
	genericclioptions.RESTClientGetter
	restConfig   *rest.Config
	clientConfig clientcmd.ClientConfig
	cache        discovery.CachedDiscoveryInterface
}
type noCacheDiscoveryClient struct {
	discovery.DiscoveryInterface
}

// Fresh is a no-op implementation of the corresponding method of the CachedDiscoveryInterface.
// No need to re-try search in the cache, return true.
func (noCacheDiscoveryClient) Fresh() bool { return true }
func (noCacheDiscoveryClient) Invalidate() {}

func NewRestGetter(restConfig *rest.Config, clientConfig clientcmd.ClientConfig) (*RestGetter, error) {
	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return &RestGetter{
		restConfig:   restConfig,
		clientConfig: clientConfig,
		cache:        noCacheDiscoveryClient{clientSet.Discovery()},
	}, nil
}

func NewRestGetterFromPath(kubeconfigPath string) (*RestGetter, error) {
	kubeConfigContent, err := os.ReadFile(kubeconfigPath) //read the content of file
	if err != nil {
		return nil, err
	}

	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeConfigContent)
	if err != nil {
		return nil, err
	}

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	cache := memory.NewMemCacheClient(discoveryClient)

	return &RestGetter{
		restConfig:   restConfig,
		clientConfig: clientConfig,
		cache:        cache,
	}, nil
}

// ToRESTConfig returns restconfig
func (r *RestGetter) ToRESTConfig() (*rest.Config, error) {
	return r.restConfig, nil
}

// ToDiscoveryClient returns a cached discovery client.
func (r *RestGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return r.cache, nil
}

// ToRESTMapper returns a RESTMapper.
func (r *RestGetter) ToRESTMapper() (meta.RESTMapper, error) {
	return restmapper.NewDeferredDiscoveryRESTMapper(r.cache), nil
}

// ToRawKubeConfigLoader return kubeconfig loader as-is
func (r *RestGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return r.clientConfig
}
