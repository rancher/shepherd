package helm

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	helmAction "helm.sh/helm/v3/pkg/action"
	helmCLI "helm.sh/helm/v3/pkg/cli"
)

func InitHelmSettings(kubeconfigPath, kubeContext, namespace string) *helmCLI.EnvSettings {
	settings := helmCLI.New()
	if kubeconfigPath != "" {
		settings.KubeConfig = kubeconfigPath
	}
	if kubeContext != "" {
		settings.KubeContext = kubeContext
	}
	if namespace != "" {
		settings.SetNamespace(namespace)
	}
	return settings
}

func InitActionConfig(settings *helmCLI.EnvSettings, namespace, helmDriver string) (*helmAction.Configuration, error) {
	if namespace == "" {
		return nil, errors.New("InitActionConfig: 'namespace' must be a non-zero length string")
	}
	actionConfig := new(helmAction.Configuration)
	if helmDriver == "" {
		helmDriver = os.Getenv("HELM_DRIVER")
	}

	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, helmDriver, func(format string, v ...interface{}) {
		_ = fmt.Sprintf(format, v)
	}); err != nil {
		return nil, err
	}
	return actionConfig, nil
}

func InitActionConfigWithGetter(getter *RESTClientGetter, namespace, helmDriver string) (*helmAction.Configuration, error) {
	actionConfig := new(helmAction.Configuration)
	if helmDriver == "" {
		helmDriver = os.Getenv("HELM_DRIVER")
	}

	if err := actionConfig.Init(getter, namespace, helmDriver, func(format string, v ...interface{}) {
		_ = fmt.Sprintf(format, v)
	}); err != nil {
		return nil, err
	}
	return actionConfig, nil
}
