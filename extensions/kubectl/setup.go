package kubectl

import (
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/kubeconfig"
	"github.com/rancher/shepherd/pkg/session"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	shepherdDynamic "github.com/rancher/shepherd/clients/dynamic"
)

func setupDynamicClient(s *session.Session, client *rancher.Client, scheme *runtime.Scheme, clusterID string) (*shepherdDynamic.Client, *session.Session, error) {
	kubeConfig, err := kubeconfig.GetKubeconfig(client, clusterID)
	if err != nil {
		return nil, s, err
	}

	restConfig, err := (*kubeConfig).ClientConfig()
	if err != nil {
		return nil, s, err
	}

	if scheme != nil {
		restConfig.ContentConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme)
	}

	var session *session.Session
	if s == nil {
		session = client.Session.NewSession()
	} else {
		session = s
	}

	dynClient, err := shepherdDynamic.NewForConfig(session, restConfig)

	return dynClient, session, err
}

func setupDynamicClientFromFlags(s *session.Session, masterURL, kubeconfigPath string, scheme *runtime.Scheme) (*shepherdDynamic.Client, *session.Session, error) {
	kubeConfig, err := kubeconfig.GetKubeconfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		return nil, s, err
	}

	restConfig, err := (*kubeConfig).ClientConfig()
	if err != nil {
		return nil, s, err
	}

	if scheme != nil {
		restConfig.ContentConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme)
	}

	var session *session.Session
	if s == nil {
		session = session.NewSession()
	} else {
		session = s
	}

	dynClient, err := shepherdDynamic.NewForConfig(session, restConfig)

	return dynClient, session, err
}
