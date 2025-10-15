package alibaba

import (
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
)

const (
	// The json/yaml config key for the ALI hosted cluster config
	ALIClusterConfigConfigurationFileKey = "aliClusterConfig"
)

// Minimal struct mapping for aliConfig JSON payload
type DataDisk struct {
	Category  string `json:"category" yaml:"category"`
	Size      int    `json:"size" yaml:"size"`
	Encrypted string `json:"encrypted" yaml:"encrypted"`
}

type NodePool struct {
	Name               string     `json:"name" yaml:"name"`
	InstanceTypes      []string   `json:"instanceTypes" yaml:"instanceTypes"`
	SystemDiskCategory string     `json:"systemDiskCategory" yaml:"systemDiskCategory"`
	SystemDiskSize     int        `json:"systemDiskSize" yaml:"systemDiskSize"`
	DataDisks          []DataDisk `json:"dataDisks" yaml:"dataDisks"`
	DesiredSize        int        `json:"desiredSize" yaml:"desiredSize"`
	ImageId            string     `json:"imageId" yaml:"imageId"`
	ImageType          string     `json:"imageType" yaml:"imageType"`
	Runtime            string     `json:"runtime" yaml:"runtime"`
	RuntimeVersion     string     `json:"runtimeVersion" yaml:"runtimeVersion"`
}

type Addon struct {
	Name string `json:"name" yaml:"name"`
}

type AliConfig struct {
	ClusterName             string     `json:"clusterName" yaml:"clusterName"`
	ClusterType             string     `json:"clusterType" yaml:"clusterType"`
	KubernetesVersion       string     `json:"kubernetesVersion" yaml:"kubernetesVersion"`
	EndpointPublicAccess    bool       `json:"endpointPublicAccess" yaml:"endpointPublicAccess"`
	Imported                bool       `json:"imported" yaml:"imported"`
	RegionId                string     `json:"regionId" yaml:"regionId"`
	ZoneIds                 []string   `json:"zoneIds" yaml:"zoneIds"`
	AlibabaCredentialSecret string     `json:"alibabaCredentialSecret" yaml:"alibabaCredentialSecret"`
	Addons                  []Addon    `json:"addons" yaml:"addons"`
	SnatEntry               bool       `json:"snatEntry" yaml:"snatEntry"`
	ServiceCidr             string     `json:"serviceCidr" yaml:"serviceCidr"`
	ResourceGroupId         string     `json:"resourceGroupId" yaml:"resourceGroupId"`
	ProxyMode               string     `json:"proxyMode" yaml:"proxyMode"`
	NodePools               []NodePool `json:"nodePools" yaml:"nodePools"`
}

type AliNodePool struct {
	Name               string            `json:"name,omitempty" yaml:"name,omitempty"`
	InstanceTypes      []string          `json:"instanceTypes,omitempty" yaml:"instanceTypes,omitempty"`
	DesiredSize        int64             `json:"desiredSize,omitempty" yaml:"desiredSize,omitempty"`
	SystemDiskCategory string            `json:"systemDiskCategory,omitempty" yaml:"systemDiskCategory,omitempty"`
	SystemDiskSize     int64             `json:"systemDiskSize,omitempty" yaml:"systemDiskSize,omitempty"`
	DataDisks          []DataDisk        `json:"dataDisks,omitempty" yaml:"dataDisks,omitempty"`
	ImageId            string            `json:"imageId,omitempty" yaml:"imageId,omitempty"`
	ImageType          string            `json:"imageType,omitempty" yaml:"imageType,omitempty"`
	Runtime            string            `json:"runtime,omitempty" yaml:"runtime,omitempty"`
	RuntimeVersion     string            `json:"runtimeVersion,omitempty" yaml:"runtimeVersion,omitempty"`
	Labels             map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Tags               map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

type ClusterConfig struct {
	Addons                    []Addon       `json:"addons,omitempty" yaml:"addons,omitempty"`
	AlibabaCredentialSecret   string        `json:"alibabaCredentialSecret,omitempty" yaml:"alibabaCredentialSecret,omitempty"`
	ClusterID                 string        `json:"clusterId,omitempty" yaml:"clusterId,omitempty"`
	ClusterName               string        `json:"clusterName,omitempty" yaml:"clusterName,omitempty"`
	ClusterSpec               string        `json:"clusterSpec,omitempty" yaml:"clusterSpec,omitempty"`
	ClusterType               string        `json:"clusterType,omitempty" yaml:"clusterType,omitempty"`
	ContainerCIDR             string        `json:"containerCidr,omitempty" yaml:"containerCidr,omitempty"`
	EndpointPublicAccess      bool          `json:"endpointPublicAccess,omitempty" yaml:"endpointPublicAccess,omitempty"`
	Imported                  bool          `json:"imported,omitempty" yaml:"imported,omitempty"`
	IsEnterpriseSecurityGroup *bool         `json:"isEnterpriseSecurityGroup,omitempty" yaml:"isEnterpriseSecurityGroup,omitempty"`
	KubernetesVersion         string        `json:"kubernetesVersion,omitempty" yaml:"kubernetesVersion,omitempty"`
	NodeCIDRMask              int64         `json:"nodeCidrMask,omitempty" yaml:"nodeCidrMask,omitempty"`
	NodePools                 []AliNodePool `json:"nodePools,omitempty" yaml:"nodePools,omitempty"`
	PodVswitchIDs             []string      `json:"podVswitchIds,omitempty" yaml:"podVswitchIds,omitempty"`
	ProxyMode                 string        `json:"proxyMode,omitempty" yaml:"proxyMode,omitempty"`
	RegionID                  string        `json:"regionId,omitempty" yaml:"regionId,omitempty"`
	ResourceGroupID           string        `json:"resourceGroupId,omitempty" yaml:"resourceGroupId,omitempty"`
	SNATEntry                 bool          `json:"snatEntry,omitempty" yaml:"snatEntry,omitempty"`
	SecurityGroupID           string        `json:"securityGroupId,omitempty" yaml:"securityGroupId,omitempty"`
	ServiceCIDR               string        `json:"serviceCidr,omitempty" yaml:"serviceCidr,omitempty"`
	VSwitchIDs                []string      `json:"vswitchIds,omitempty" yaml:"vswitchIds,omitempty"`
	VpcID                     string        `json:"vpcId,omitempty" yaml:"vpcId,omitempty"`
	ZoneIDs                   []string      `json:"zoneIds,omitempty" yaml:"zoneIds,omitempty"`
}

// aliHostClusterConfig maps AliConfig to Rancher AliClusterConfigSpec
func aliHostClusterConfig(displayName, cloudCredentialID string, aliClusterConfig ClusterConfig) *management.AliClusterConfigSpec {
	return &management.AliClusterConfigSpec{
		AlibabaCredentialSecret: cloudCredentialID,
		ClusterName:             displayName,
		ClusterType:             aliClusterConfig.ClusterType,
		KubernetesVersion:       aliClusterConfig.KubernetesVersion,
		EndpointPublicAccess:    aliClusterConfig.EndpointPublicAccess,
		Imported:                aliClusterConfig.Imported,
		RegionID:                aliClusterConfig.RegionID,
		ZoneIDs:                 aliClusterConfig.ZoneIDs,
		Addons:                  mapAliAddons(aliClusterConfig.Addons),
		SNATEntry:               aliClusterConfig.SNATEntry,
		ServiceCIDR:             aliClusterConfig.ServiceCIDR,
		ResourceGroupID:         aliClusterConfig.ResourceGroupID,
		ProxyMode:               aliClusterConfig.ProxyMode,
		NodePools:               MapAliNodePoolsFromAliNodePool(aliClusterConfig.NodePools),
	}
}

// Helper to map Addons
func mapAliAddons(addons []Addon) []management.AliAddon {
	out := make([]management.AliAddon, len(addons))
	for i, a := range addons {
		out[i] = management.AliAddon{Name: a.Name}
	}
	return out
}

// MapAliNodePoolsFromAliNodePool maps []AliNodePool to []management.AliNodePool (exported)
func MapAliNodePoolsFromAliNodePool(pools []AliNodePool) []management.AliNodePool {
	out := make([]management.AliNodePool, len(pools))
	for i, p := range pools {
		out[i] = management.AliNodePool{
			Name:               p.Name,
			InstanceTypes:      p.InstanceTypes,
			DesiredSize:        int64Ptr(p.DesiredSize),
			SystemDiskCategory: p.SystemDiskCategory,
			SystemDiskSize:     p.SystemDiskSize,
			ImageID:            p.ImageId,
			ImageType:          p.ImageType,
			Runtime:            p.Runtime,
			RuntimeVersion:     p.RuntimeVersion,
			DataDisks:          MapAliDataDisks(p.DataDisks),
		}
	}
	return out
}

// MapAliDataDisks converts []DataDisk to []management.AliDisk
func MapAliDataDisks(disks []DataDisk) []management.AliDisk {
	out := make([]management.AliDisk, len(disks))
	for i, d := range disks {
		out[i] = management.AliDisk{Category: d.Category, Size: int64(d.Size), Encrypted: d.Encrypted}
	}
	return out
}

// int64Ptr returns a pointer to an int64 value.
func int64Ptr(i int64) *int64 {
	return &i
}
