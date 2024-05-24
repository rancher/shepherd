package helm

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type RESTClientGetter struct {
	genericclioptions.RESTClientGetter
	config       *genericclioptions.ConfigFlags
	restConfig   *rest.Config
	clientConfig clientcmd.ClientConfig
	cache        discovery.CachedDiscoveryInterface
}

func NewRESTClientGetter(restConfig *rest.Config, clientConfig *clientcmd.ClientConfig, namespace string) (*RESTClientGetter, error) {
	if restConfig == nil || clientConfig == nil {
		return nil, errors.New("NewRESTClientGetter: 'restConfig' and 'clientConfig' must be non-nil")
	}

	// namespace must be overridden if modifying/deploying things in a non-default namespace
	if namespace != "" {
		rawConfig, err := (*clientConfig).RawConfig()
		if err != nil {
			return nil, err
		}

		config := clientcmd.NewDefaultClientConfig(rawConfig, &clientcmd.ConfigOverrides{
			Context: api.Context{
				Namespace: namespace,
			},
		})
		clientConfig = &config
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	cache := memory.NewMemCacheClient(discoveryClient)
	return &RESTClientGetter{
		config:       genericclioptions.NewConfigFlags(false),
		restConfig:   restConfig,
		clientConfig: *clientConfig,
		cache:        cache,
	}, nil
}

func NewRESTClientGetterFromBytes(kubeConfigContent []byte, namespace string) (*RESTClientGetter, error) {
	if len(kubeConfigContent) < 1 {
		return nil, errors.New("NewRESTClientGetterFromBytes: 'kubeConfigContent' must be a non-zero length []byte")
	}
	if namespace == "" {
		return nil, errors.New("NewRESTClientGetterFromBytes: 'namespace' must be a non-zero length string")
	}
	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeConfigContent)
	if err != nil {
		return nil, err
	}

	rawconfig, err := clientConfig.RawConfig()
	if err != nil {
		return nil, err
	}

	clientConfig = clientcmd.NewDefaultClientConfig(rawconfig, &clientcmd.ConfigOverrides{
		Context: api.Context{
			Namespace: namespace,
		},
	})

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	cache := memory.NewMemCacheClient(discoveryClient)

	return &RESTClientGetter{
		restConfig:   restConfig,
		clientConfig: clientConfig,
		cache:        cache,
	}, nil
}

// ToRESTConfig returns restconfig
func (r *RESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	return r.restConfig, nil
}

// ToDiscoveryClient returns a cached discovery client.
func (r *RESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return r.cache, nil
}

// ToRESTMapper returns a RESTMapper.
func (r *RESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	return restmapper.NewDeferredDiscoveryRESTMapper(r.cache), nil
}

// ToRawKubeConfigLoader return kubeconfig loader as-is
func (r *RESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return r.clientConfig
}
