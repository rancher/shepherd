package charts

import (
	"context"
	"fmt"

	catalogv1 "github.com/rancher/rancher/pkg/apis/catalog.cattle.io/v1"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/clients/rancher/catalog"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/pkg/api/steve/catalog/types"
	"github.com/rancher/shepherd/pkg/wait"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	// Namespace that rancher CIS Benchmark chart is installed in
	RancherCisBenchmarkNamespace = "cis-operator-system"
	// Name of the rancher CIS Benchmark chart
	RancherCisBenchmarkName = "rancher-cis-benchmark"
	// Name of rancher CIS Benchmark crd chart
	RancherCisBenchmarkCRDName = "rancher-cis-benchmark-crd"
	CisbenchmarkProjectName    = "cis-operator-system"
	ClusterScanResourceType    = "cis.cattle.io.clusterscan"
	ClusterScanReportType      = "cis.cattle.io.clusterscanreport"
)

// ClusterScanStatus represents the status field of a cluster scan object.
type ClusterScanStatus struct {
	Display Display
}

// Display contains the state of the cluster scan.
// State can be pending, running, reporting, pass, and fail
type Display struct {
	State   string
	Message string
}

// ClusterScanReportSpec represents the specification for a cluster scan report.
type ClusterScanReportSpec struct {
	ReportJson string
}

// CisReport is the report structure stored as report json in cluster scan report spec.
type CisReport struct {
	Total         int
	Pass          int
	Fail          int
	Skip          int
	Warn          int
	NotApplicable int
	Results       []*Group `json:"results"`
}

// Group is the result structure stored as report json in Results of CisReport
type Group struct {
	ID     string      `yaml:"id" json:"id"`
	Text   string      `json:"description"`
	Checks []*CisCheck `json:"checks"`
}

// CisCheck is the ID, Description and State structure of individual test in cluster scan.
type CisCheck struct {
	Id          string
	Description string
	State       string
}

// InstallRancherCisBenchmarkChart is a helper function that installs the rancher-CIS Benchmark chart.
func InstallRancherCisBenchmarkChart(client *rancher.Client, installOptions *InstallOptions) error {
	serverSetting, err := client.Management.Setting.ByID(serverURLSettingID)
	if err != nil {
		return err
	}

	registrySetting, err := client.Management.Setting.ByID(defaultRegistrySettingID)
	if err != nil {
		return err
	}

	cisbenchmarkChartInstallActionPayload := &payloadOpts{
		InstallOptions:  *installOptions,
		Name:            RancherCisBenchmarkName,
		Namespace:       RancherCisBenchmarkNamespace,
		Host:            serverSetting.Value,
		DefaultRegistry: registrySetting.Value,
	}

	chartInstallAction := newCisBenchmarkChartInstallAction(cisbenchmarkChartInstallActionPayload)

	catalogClient, err := client.GetClusterCatalogClient(installOptions.Cluster.ID)
	if err != nil {
		return err
	}

	// Install the Rancher CIS Benchmark chart
	err = catalogClient.InstallChart(chartInstallAction, catalog.RancherChartRepo)
	if err != nil {
		return err
	}

	// Wait for chart to be fully deployed
	watchAppInterface, err := catalogClient.Apps(RancherCisBenchmarkNamespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector:  "metadata.name=" + RancherCisBenchmarkName,
		TimeoutSeconds: &defaults.WatchTimeoutSeconds,
	})
	if err != nil {
		return err
	}

	err = wait.WatchWait(watchAppInterface, func(event watch.Event) (ready bool, err error) {
		app := event.Object.(*catalogv1.App)

		state := app.Status.Summary.State
		if state == string(catalogv1.StatusDeployed) {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}

	// Register cleanup function for uninstallation
	client.Session.RegisterCleanupFunc(func() error {
		return UninstallRancherCisBenchmarkChart(client, installOptions)
	})

	return nil
}

// UpgradeRanchercisbenchamrkChart is a helper function that upgrades the rancher-cisbenchmark chart.
func UpgradeRancherCisBenchmarkChart(client *rancher.Client, installOptions *InstallOptions) error {
	serverSetting, err := client.Management.Setting.ByID(serverURLSettingID)
	if err != nil {
		return err
	}

	registrySetting, err := client.Management.Setting.ByID(defaultRegistrySettingID)
	if err != nil {
		return err
	}

	cisbenchmarkChartUpgradeActionPayload := &payloadOpts{
		InstallOptions:  *installOptions,
		Name:            RancherCisBenchmarkName,
		Namespace:       RancherCisBenchmarkNamespace,
		Host:            serverSetting.Value,
		DefaultRegistry: registrySetting.Value,
	}

	chartUpgradeAction := newCisBenchmarkChartUpgradeAction(cisbenchmarkChartUpgradeActionPayload)

	catalogClient, err := client.GetClusterCatalogClient(installOptions.Cluster.ID)
	if err != nil {
		return err
	}

	err = catalogClient.UpgradeChart(chartUpgradeAction, catalog.RancherChartRepo)
	if err != nil {
		return err
	}

	adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
	if err != nil {
		return err
	}
	adminCatalogClient, err := adminClient.GetClusterCatalogClient(installOptions.Cluster.ID)
	if err != nil {
		return err
	}

	// wait for chart to be in status pending upgrade
	watchAppInterface, err := adminCatalogClient.Apps(RancherCisBenchmarkNamespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector:  "metadata.name=" + RancherCisBenchmarkName,
		TimeoutSeconds: &defaults.WatchTimeoutSeconds,
	})
	if err != nil {
		return err
	}

	err = wait.WatchWait(watchAppInterface, func(event watch.Event) (ready bool, err error) {
		app := event.Object.(*catalogv1.App)

		state := app.Status.Summary.State
		if state == string(catalogv1.StatusPendingUpgrade) {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}

	// wait for chart to be full deployed
	watchAppInterface, err = adminCatalogClient.Apps(RancherCisBenchmarkNamespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector:  "metadata.name=" + RancherCisBenchmarkName,
		TimeoutSeconds: &defaults.WatchTimeoutSeconds,
	})
	if err != nil {
		return err
	}

	err = wait.WatchWait(watchAppInterface, func(event watch.Event) (ready bool, err error) {
		app := event.Object.(*catalogv1.App)

		state := app.Status.Summary.State
		if state == string(catalogv1.StatusDeployed) {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// newCisBenchmarkChartUpgradeAction is a private helper function that returns chart upgrade action.
func newCisBenchmarkChartUpgradeAction(p *payloadOpts) *types.ChartUpgradeAction {
	chartUpgrade := newChartUpgrade(p.Name, p.Name, p.InstallOptions.Version, p.InstallOptions.Cluster.ID, p.InstallOptions.Cluster.Name, p.Host, p.DefaultRegistry, nil)
	chartUpgradeCRD := newChartUpgrade(p.Name+"-crd", p.Name+"-crd", p.InstallOptions.Version, p.InstallOptions.Cluster.ID, p.InstallOptions.Cluster.Name, p.Host, p.DefaultRegistry, nil)
	chartUpgrades := []types.ChartUpgrade{*chartUpgradeCRD, *chartUpgrade}

	chartUpgradeAction := newChartUpgradeAction(p.Namespace, chartUpgrades)

	return chartUpgradeAction
}

// UninstallRancherCisBenchmarkChart is a helper function that uninstalls the rancher-CIS Benchmark chart.
func UninstallRancherCisBenchmarkChart(client *rancher.Client, installOptions *InstallOptions) error {
	catalogClient, err := client.GetClusterCatalogClient(installOptions.Cluster.ID)
	if err != nil {
		return err
	}

	// Uninstall the Rancher CIS Benchmark chart
	defaultChartUninstallAction := newChartUninstallAction()
	err = catalogClient.UninstallChart(RancherCisBenchmarkName, RancherCisBenchmarkNamespace, defaultChartUninstallAction)
	if err != nil {
		return err
	}

	// Watch for events related to the uninstallation and wait until it is deleted
	watchAppInterface, err := catalogClient.Apps(RancherCisBenchmarkNamespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector:  "metadata.name=" + RancherCisBenchmarkName,
		TimeoutSeconds: &defaults.WatchTimeoutSeconds,
	})
	if err != nil {
		return err
	}

	err = wait.WatchWait(watchAppInterface, func(event watch.Event) (ready bool, err error) {
		chart := event.Object.(*catalogv1.App)
		if event.Type == watch.Error {
			return false, fmt.Errorf("there was an error uninstalling rancher CIS Benchmark chart")
		} else if event.Type == watch.Deleted {
			return true, nil
		} else if chart == nil {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}

	// Uninstall the Rancher CIS Benchmark CRD chart
	err = catalogClient.UninstallChart(RancherCisBenchmarkCRDName, RancherCisBenchmarkNamespace, defaultChartUninstallAction)
	if err != nil {
		return err
	}

	// Watch for events related to the uninstallation and wait until it is deleted
	watchAppInterface, err = catalogClient.Apps(RancherCisBenchmarkNamespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector:  "metadata.name=" + RancherCisBenchmarkCRDName,
		TimeoutSeconds: &defaults.WatchTimeoutSeconds,
	})
	if err != nil {
		return err
	}

	err = wait.WatchWait(watchAppInterface, func(event watch.Event) (ready bool, err error) {
		chart := event.Object.(*catalogv1.App)
		if event.Type == watch.Error {
			return false, fmt.Errorf("there was an error uninstalling rancher CIS Benchmark chart")
		} else if event.Type == watch.Deleted {
			return true, nil
		} else if chart == nil {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// newCisBenchmarkChartInstallAction is a helper function that returns an array of newChartInstallActions for installing the cis-benchmark and cis-benchmark-crd charts
func newCisBenchmarkChartInstallAction(p *payloadOpts) *types.ChartInstallAction {
	chartInstall := newChartInstall(p.Name, p.InstallOptions.Version, p.InstallOptions.Cluster.ID, p.InstallOptions.Cluster.Name, p.Host, rancherChartsName, p.ProjectID, p.DefaultRegistry, nil)
	chartInstallCRD := newChartInstall(p.Name+"-crd", p.InstallOptions.Version, p.InstallOptions.Cluster.ID, p.InstallOptions.Cluster.Name, p.Host, rancherChartsName, p.ProjectID, p.DefaultRegistry, nil)

	chartInstalls := []types.ChartInstall{*chartInstallCRD, *chartInstall}

	chartInstallAction := newChartInstallAction(p.Namespace, p.ProjectID, chartInstalls)

	return chartInstallAction
}
