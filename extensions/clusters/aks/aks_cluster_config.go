package aks

import (
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
)

const (
	// The json/yaml config key for the AKS hosted cluster config
	AKSClusterConfigConfigurationFileKey = "aksClusterConfig"
)

// ClusterConfig is the configuration needed to create an AKS host cluster
type ClusterConfig struct {
	AuthBaseURL                 *string           `json:"authBaseUrl,omitempty" yaml:"authBaseUrl,omitempty"`
	AuthorizedIPRanges          *[]string         `json:"authorizedIpRanges,omitempty" yaml:"authorizedIpRanges,omitempty"`
	AzureCredentialSecret       string            `json:"azureCredentialSecret" yaml:"azureCredentialSecret"`
	BaseURL                     *string           `json:"baseUrl,omitempty" yaml:"baseUrl,omitempty"`
	DNSPrefix                   *string           `json:"dnsPrefix,omitempty" yaml:"dnsPrefix,omitempty"`
	HTTPApplicationRouting      *bool             `json:"httpApplicationRouting,omitempty" yaml:"httpApplicationRouting,omitempty"`
	KubernetesVersion           *string           `json:"kubernetesVersion,omitempty" yaml:"kubernetesVersion,omitempty"`
	LinuxAdminUsername          *string           `json:"linuxAdminUsername,omitempty" yaml:"linuxAdminUsername,omitempty"`
	LinuxSSHPublicKey           *string           `json:"sshPublicKey,omitempty" yaml:"sshPublicKey,omitempty"`
	LoadBalancerSKU             *string           `json:"loadBalancerSku,omitempty" yaml:"loadBalancerSku,omitempty"`
	LogAnalyticsWorkspaceGroup  *string           `json:"logAnalyticsWorkspaceGroup,omitempty" yaml:"logAnalyticsWorkspaceGroup,omitempty"`
	LogAnalyticsWorkspaceName   *string           `json:"logAnalyticsWorkspaceName,omitempty" yaml:"logAnalyticsWorkspaceName,omitempty"`
	ManagedIdentity             **bool            `json:"managedIdentity,omitempty" yaml:"managedIdentity,omitempty"`
	Monitoring                  *bool             `json:"monitoring,omitempty" yaml:"monitoring,omitempty"`
	NetworkDNSServiceIP         *string           `json:"dnsServiceIp,omitempty" yaml:"dnsServiceIp,omitempty"`
	NetworkDockerBridgeCIDR     *string           `json:"dockerBridgeCidr,omitempty" yaml:"dockerBridgeCidr,omitempty"`
	NetworkPlugin               *string           `json:"networkPlugin,omitempty" yaml:"networkPlugin,omitempty"`
	NetworkPodCIDR              *string           `json:"podCidr,omitempty" yaml:"podCidr,omitempty"`
	NetworkPolicy               *string           `json:"networkPolicy,omitempty" yaml:"networkPolicy,omitempty"`
	NetworkServiceCIDR          *string           `json:"serviceCidr,omitempty" yaml:"serviceCidr,omitempty"`
	NodePools                   *[]NodePool       `json:"nodePools,omitempty" yaml:"nodePools,omitempty"`
	NodeResourceGroup           *string           `json:"nodeResourceGroup,omitempty" yaml:"nodeResourceGroup,omitempty"`
	OutboundType                *string           `json:"outboundType,omitempty" yaml:"outboundType,omitempty"`
	PrivateCluster              *bool             `json:"privateCluster,omitempty" yaml:"privateCluster,omitempty"`
	PrivateDNSZone              *string           `json:"privateDnsZone,omitempty" yaml:"privateDnsZone,omitempty"`
	ResourceGroup               string            `json:"resourceGroup" yaml:"resourceGroup"`
	ResourceLocation            string            `json:"resourceLocation" yaml:"resourceLocation"`
	Subnet                      *string           `json:"subnet,omitempty" yaml:"subnet,omitempty"`
	Tags                        map[string]string `json:"tags" yaml:"tags"`
	UserAssignedIdentity        *string           `json:"userAssignedIdentity,omitempty" yaml:"userAssignedIdentity,omitempty"`
	VirtualNetwork              *string           `json:"virtualNetwork,omitempty" yaml:"virtualNetwork,omitempty"`
	VirtualNetworkResourceGroup *string           `json:"virtualNetworkResourceGroup,omitempty" yaml:"virtualNetworkResourceGroup,omitempty"`
}

// NodePool is the configuration needed to an AKS node pool
type NodePool struct {
	AvailabilityZones   *[]string         `json:"availabilityZones,omitempty" yaml:"availabilityZones,omitempty"`
	EnableAutoScaling   *bool             `json:"enableAutoScaling,omitempty" yaml:"enableAutoScaling,omitempty"`
	MaxPods             *int64            `json:"maxPods,omitempty" yaml:"maxPods,omitempty"`
	MaxSurge            string            `json:"maxSurge,omitempty" yaml:"maxSurge,omitempty"`
	MaxCount            *int64            `json:"maxCount,omitempty" yaml:"maxCount,omitempty"`
	MinCount            *int64            `json:"minCount,omitempty" yaml:"minCount,omitempty"`
	Mode                string            `json:"mode" yaml:"mode"`
	Name                *string           `json:"name,omitempty" yaml:"name,omitempty"`
	NodeCount           *int64            `json:"nodeCount,omitempty" yaml:"nodeCount,omitempty"`
	NodeLabels          map[string]string `json:"nodeLabels,omitempty" yaml:"nodeLabels,omitempty"`
	NodeTaints          []string          `json:"nodeTaints,omitempty" yaml:"nodeTaints,omitempty"`
	OrchestratorVersion *string           `json:"orchestratorVersion,omitempty" yaml:"orchestratorVersion,omitempty"`
	OsDiskSizeGB        *int64            `json:"osDiskSizeGB,omitempty" yaml:"osDiskSizeGB,omitempty"`
	OsDiskType          string            `json:"osDiskType" yaml:"osDiskType"`
	OsType              string            `json:"osType" yaml:"osType"`
	VMSize              string            `json:"vmSize" yaml:"vmSize"`
	VnetSubnetID        *string           `json:"vnetSubnetID,omitempty" yaml:"vnetSubnetID,omitempty"`
}

func aksNodePoolConstructor(aksNodePoolConfigs *[]NodePool, kubernetesVersion string) *[]management.AKSNodePool {
	var aksNodePools = make([]management.AKSNodePool, 0)
	if aksNodePoolConfigs == nil {
		return nil
	}
	for _, aksNodePoolConfig := range *aksNodePoolConfigs {
		aksNodePool := management.AKSNodePool{
			AvailabilityZones:   aksNodePoolConfig.AvailabilityZones,
			Count:               aksNodePoolConfig.NodeCount,
			EnableAutoScaling:   aksNodePoolConfig.EnableAutoScaling,
			MaxCount:            aksNodePoolConfig.MaxCount,
			MaxPods:             aksNodePoolConfig.MaxPods,
			MaxSurge:            aksNodePoolConfig.MaxSurge,
			MinCount:            aksNodePoolConfig.MinCount,
			Mode:                aksNodePoolConfig.Mode,
			Name:                aksNodePoolConfig.Name,
			NodeLabels:          aksNodePoolConfig.NodeLabels,
			NodeTaints:          aksNodePoolConfig.NodeTaints,
			OrchestratorVersion: &kubernetesVersion,
			OsDiskSizeGB:        aksNodePoolConfig.OsDiskSizeGB,
			OsDiskType:          aksNodePoolConfig.OsDiskType,
			OsType:              aksNodePoolConfig.OsType,
			VMSize:              aksNodePoolConfig.VMSize,
			VnetSubnetID:        aksNodePoolConfig.VnetSubnetID,
		}
		aksNodePools = append(aksNodePools, aksNodePool)
	}
	return &aksNodePools
}

func HostClusterConfig(displayName, cloudCredentialID string, aksClusterConfig ClusterConfig) *management.AKSClusterConfigSpec {
	return &management.AKSClusterConfigSpec{
		AuthBaseURL:                 aksClusterConfig.AuthBaseURL,
		AuthorizedIPRanges:          aksClusterConfig.AuthorizedIPRanges,
		AzureCredentialSecret:       cloudCredentialID,
		BaseURL:                     aksClusterConfig.BaseURL,
		ClusterName:                 displayName,
		DNSPrefix:                   aksClusterConfig.DNSPrefix,
		HTTPApplicationRouting:      aksClusterConfig.HTTPApplicationRouting,
		Imported:                    false,
		KubernetesVersion:           aksClusterConfig.KubernetesVersion,
		LinuxAdminUsername:          aksClusterConfig.LinuxAdminUsername,
		LinuxSSHPublicKey:           aksClusterConfig.LinuxSSHPublicKey,
		LoadBalancerSKU:             aksClusterConfig.LoadBalancerSKU,
		LogAnalyticsWorkspaceGroup:  aksClusterConfig.LogAnalyticsWorkspaceGroup,
		LogAnalyticsWorkspaceName:   aksClusterConfig.LogAnalyticsWorkspaceName,
		ManagedIdentity:             aksClusterConfig.ManagedIdentity,
		Monitoring:                  aksClusterConfig.Monitoring,
		NetworkDNSServiceIP:         aksClusterConfig.NetworkDNSServiceIP,
		NetworkDockerBridgeCIDR:     aksClusterConfig.NetworkDockerBridgeCIDR,
		NetworkPlugin:               aksClusterConfig.NetworkPlugin,
		NetworkPodCIDR:              aksClusterConfig.NetworkPodCIDR,
		NetworkPolicy:               aksClusterConfig.NetworkPolicy,
		NetworkServiceCIDR:          aksClusterConfig.NetworkServiceCIDR,
		NodePools:                   aksNodePoolConstructor(aksClusterConfig.NodePools, *aksClusterConfig.KubernetesVersion),
		NodeResourceGroup:           aksClusterConfig.NodeResourceGroup,
		OutboundType:                aksClusterConfig.OutboundType,
		PrivateCluster:              aksClusterConfig.PrivateCluster,
		PrivateDNSZone:              aksClusterConfig.PrivateDNSZone,
		ResourceGroup:               aksClusterConfig.ResourceGroup,
		ResourceLocation:            aksClusterConfig.ResourceLocation,
		Subnet:                      aksClusterConfig.Subnet,
		Tags:                        aksClusterConfig.Tags,
		UserAssignedIdentity:        aksClusterConfig.UserAssignedIdentity,
		VirtualNetwork:              aksClusterConfig.VirtualNetwork,
		VirtualNetworkResourceGroup: aksClusterConfig.VirtualNetworkResourceGroup,
	}
}
