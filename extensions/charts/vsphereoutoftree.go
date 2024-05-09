package charts

import (
	"github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	"github.com/rancher/shepherd/extensions/projects"
	"github.com/rancher/shepherd/extensions/rke1/nodetemplates"
	"github.com/rancher/shepherd/pkg/api/steve/catalog/types"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/clients/rancher/catalog"
	r1vsphere "github.com/rancher/shepherd/extensions/rke1/nodetemplates/vsphere"
)

const (
	systemProject       = "System"
	vsphereCPIchartName = "rancher-vsphere-cpi"
	vsphereCSIchartName = "rancher-vsphere-csi"

	vcenter      = "vCenter"
	storageclass = "storageClass"

	datacenters  = "datacenters"
	host         = "host"
	password     = "password"
	username     = "username"
	port         = "port"
	clusterid    = "clusterId"
	datastoreurl = "datastoreURL"
)

// InstallVsphereOutOfTreeCharts installs the CPI and CSI chart for aws cloud provider in a given cluster.
func InstallVsphereOutOfTreeCharts(client *rancher.Client, vsphereTemplate *nodetemplates.NodeTemplate, repoName, clusterName string) error {
	serverSetting, err := client.Management.Setting.ByID(serverURLSettingID)
	if err != nil {
		return err
	}

	cluster, err := clusters.NewClusterMeta(client, clusterName)
	if err != nil {
		return err
	}

	registrySetting, err := client.Management.Setting.ByID(defaultRegistrySettingID)
	if err != nil {
		return err
	}

	project, err := projects.GetProjectByName(client, cluster.ID, systemProject)
	if err != nil {
		return err
	}

	catalogClient, err := client.GetClusterCatalogClient(cluster.ID)
	if err != nil {
		return err
	}

	latestCPIVersion, err := catalogClient.GetLatestChartVersion(vsphereCPIchartName, catalog.RancherChartRepo)
	if err != nil {
		return err
	}

	installCPIOptions := &InstallOptions{
		Cluster:   cluster,
		Version:   latestCPIVersion,
		ProjectID: project.ID,
	}

	chartInstallActionPayload := &payloadOpts{
		InstallOptions:  *installCPIOptions,
		Name:            vsphereCPIchartName,
		Namespace:       namespaces.KubeSystem,
		Host:            serverSetting.Value,
		DefaultRegistry: registrySetting.Value,
	}

	chartInstallAction, err := vsphereCPIChartInstallAction(catalogClient,
		chartInstallActionPayload, vsphereTemplate, installCPIOptions, repoName, namespaces.KubeSystem)
	if err != nil {
		return err
	}

	err = catalogClient.InstallChart(chartInstallAction, repoName)
	if err != nil {
		return err
	}

	err = VerifyChartInstall(catalogClient, namespaces.KubeSystem, vsphereCPIchartName)
	if err != nil {
		return err
	}

	latestCSIVersion, err := catalogClient.GetLatestChartVersion(vsphereCSIchartName, catalog.RancherChartRepo)
	if err != nil {
		return err
	}

	installCSIOptions := &InstallOptions{
		Cluster:   cluster,
		Version:   latestCSIVersion,
		ProjectID: project.ID,
	}

	chartInstallActionPayload = &payloadOpts{
		InstallOptions:  *installCSIOptions,
		Name:            vsphereCSIchartName,
		Namespace:       namespaces.KubeSystem,
		Host:            serverSetting.Value,
		DefaultRegistry: registrySetting.Value,
	}

	chartInstallAction, err = vsphereCSIChartInstallAction(catalogClient, chartInstallActionPayload,
		vsphereTemplate, installCSIOptions, repoName, namespaces.KubeSystem)
	if err != nil {
		return err
	}

	err = catalogClient.InstallChart(chartInstallAction, repoName)
	if err != nil {
		return err
	}

	return err
}

// vsphereCPIChartInstallAction is a helper function that returns a chartInstallAction for aws out-of-tree chart.
func vsphereCPIChartInstallAction(client *catalog.Client, chartInstallActionPayload *payloadOpts, vsphereTemplate *nodetemplates.NodeTemplate, installOptions *InstallOptions, repoName, chartNamespace string) (*types.ChartInstallAction, error) {
	chartValues, err := client.GetChartValues(repoName, vsphereCPIchartName, installOptions.Version)
	if err != nil {
		return nil, err
	}

	chartValues[vcenter].(map[string]interface{})[datacenters] = vsphereTemplate.VmwareVsphereNodeTemplateConfig.Datacenter
	chartValues[vcenter].(map[string]interface{})[host] = vsphereTemplate.VmwareVsphereNodeTemplateConfig.Vcenter
	chartValues[vcenter].(map[string]interface{})[password] = r1vsphere.GetVspherePassword()
	chartValues[vcenter].(map[string]interface{})[username] = vsphereTemplate.VmwareVsphereNodeTemplateConfig.Username
	chartValues[vcenter].(map[string]interface{})[port] = vsphereTemplate.VmwareVsphereNodeTemplateConfig.VcenterPort

	chartInstall := newChartInstall(
		chartInstallActionPayload.Name,
		chartInstallActionPayload.Version,
		chartInstallActionPayload.Cluster.ID,
		chartInstallActionPayload.Cluster.Name,
		chartInstallActionPayload.Host,
		repoName,
		installOptions.ProjectID,
		chartInstallActionPayload.DefaultRegistry,
		chartValues)
	chartInstalls := []types.ChartInstall{*chartInstall}

	return newChartInstallAction(chartNamespace, chartInstallActionPayload.ProjectID, chartInstalls), nil
}

// vsphereCSIChartInstallAction is a helper function that returns a chartInstallAction for aws out-of-tree chart.
func vsphereCSIChartInstallAction(client *catalog.Client, chartInstallActionPayload *payloadOpts, vsphereTemplate *nodetemplates.NodeTemplate, installOptions *InstallOptions, repoName, chartNamespace string) (*types.ChartInstallAction, error) {
	chartValues, err := client.GetChartValues(repoName, vsphereCSIchartName, installOptions.Version)
	if err != nil {
		return nil, err
	}

	chartValues[vcenter].(map[string]interface{})[datacenters] = vsphereTemplate.VmwareVsphereNodeTemplateConfig.Datacenter
	chartValues[vcenter].(map[string]interface{})[host] = vsphereTemplate.VmwareVsphereNodeTemplateConfig.Vcenter
	chartValues[vcenter].(map[string]interface{})[password] = r1vsphere.GetVspherePassword()
	chartValues[vcenter].(map[string]interface{})[username] = vsphereTemplate.VmwareVsphereNodeTemplateConfig.Username
	chartValues[vcenter].(map[string]interface{})[port] = vsphereTemplate.VmwareVsphereNodeTemplateConfig.VcenterPort
	chartValues[vcenter].(map[string]interface{})[clusterid] = installOptions.Cluster.ID

	chartValues[storageclass].(map[string]interface{})[datastoreurl] = r1vsphere.GetVsphereDatastoreURL()

	chartInstall := newChartInstall(
		chartInstallActionPayload.Name,
		chartInstallActionPayload.Version,
		chartInstallActionPayload.Cluster.ID,
		chartInstallActionPayload.Cluster.Name,
		chartInstallActionPayload.Host,
		repoName,
		installOptions.ProjectID,
		chartInstallActionPayload.DefaultRegistry,
		chartValues)
	chartInstalls := []types.ChartInstall{*chartInstall}

	return newChartInstallAction(chartNamespace, chartInstallActionPayload.ProjectID, chartInstalls), nil
}
