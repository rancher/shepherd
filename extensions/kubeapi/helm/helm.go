package helm

import (
	"context"

	"helm.sh/helm/v3/pkg/release"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/rancher/shepherd/clients/helm"
	"github.com/rancher/shepherd/extensions/kubeconfig"
	"github.com/rancher/shepherd/pkg/session"
)

const (
	certManagerRepoURL          = "https://charts.jetstack.io"
	certManagerReleaseNamespace = "cert-manager"
	certManagerChartRef         = "jetstack/cert-manager"
	stableRancherRepoName       = "rancher-stable"
	stableRancherRepoURL        = "https://releases.rancher.com/server-charts/stable"
	latestRancherRepoName       = "rancher-latest"
	latestRancherRepoURL        = "https://releases.rancher.com/server-charts/latest"
	alphaRancherRepoName        = "rancher-alpha"
	alphaRancherRepoURL         = "https://releases.rancher.com/server-charts/alpha"
	rancherReleaseName          = "rancher"
	rancherNamespace            = "cattle-system"
	rancherLocalCluster         = "local"
	rancherCleanupURL           = "https://github.com/rancher/rancher-cleanup/blob/main/deploy/rancher-cleanup.yaml"
	rancherCleanupVerifyURL     = "https://github.com/rancher/rancher-cleanup/blob/main/deploy/verify.yaml"
)

// InstallRancher installs latest version of rancher including cert-manager
// using helm CLI with some predefined values set such as
// - BootstrapPassword : admin
// - Hostname          : Localhost
// - BundledMode       : True
// - Replicas          : 1
func InstallRancher(ts *session.Session, restConfig *rest.Config, rancherVersion, certManagerVersion string, vals map[string]interface{}) (*release.Release, error) {
	// if ts != nil {
	// 	ts.RegisterCleanupFunc(func() error {
	// 		kubectl.CreateUnstructured()

	// 	})
	// }
	//  ClientSet of kubernetes
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	// Create namespace cattle-system
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: rancherNamespace}}
	_, err = clientset.CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	// Install cert-manager chart
	_, err = InstallCertManager(ts, restConfig, certManagerVersion)
	if err != nil {
		return nil, err
	}

	kubeconfigBytes, err := kubeconfig.GenerateKubeconfigForRestConfig(restConfig, rancherLocalCluster, rancherLocalCluster, rancherLocalCluster)
	if err != nil {
		return nil, err
	}
	getter, err := helm.NewRESTClientGetterFromBytes(kubeconfigBytes, rancherNamespace)
	if err != nil {
		return nil, err
	}
	settings := helm.InitHelmSettings(string(kubeconfigBytes), rancherLocalCluster, rancherNamespace)

	helmClient, err := helm.NewClient(ts, settings, getter, rancherNamespace, "")
	if err != nil {
		return nil, err
	}

	// Add Rancher helm repo
	err = helmClient.AddOrUpdateRepo(stableRancherRepoName, stableRancherRepoURL, false)
	if err != nil {
		return nil, err
	}

	// Install Rancher Chart
	return helmClient.InstallChart(rancherReleaseName,
		stableRancherRepoName+"/"+rancherReleaseName,
		stableRancherRepoURL,
		rancherNamespace,
		rancherVersion,
		false,
		vals,
	)
}

// InstallCertManager installs latest version cert manager available through helm
// CLI. It sets the installCRDs as true to install crds as well.
func InstallCertManager(ts *session.Session, restConfig *rest.Config, version string) (*release.Release, error) {
	//  ClientSet of kubernetes
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	// Create namespace cert-manager
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: certManagerReleaseNamespace}}
	_, err = clientset.CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	kubeconfigBytes, err := kubeconfig.GenerateKubeconfigForRestConfig(restConfig, rancherLocalCluster, rancherLocalCluster, rancherLocalCluster)
	if err != nil {
		return nil, err
	}
	getter, err := helm.NewRESTClientGetterFromBytes(kubeconfigBytes, certManagerReleaseNamespace)
	if err != nil {
		return nil, err
	}
	helmClient, err := helm.NewClient(ts, nil, getter, certManagerReleaseNamespace, "")
	if err != nil {
		return nil, err
	}

	// Add cert-manager Helm Repo
	err = helmClient.AddOrUpdateRepo("jetstack", certManagerRepoURL, false)
	if err != nil {
		return nil, err
	}

	vals := map[string]interface{}{"installCRDs": true}
	// Install cert-manager Chart
	return helmClient.InstallChart(certManagerReleaseNamespace, "jetstack", certManagerRepoURL, certManagerReleaseNamespace, version, false, vals)
}
