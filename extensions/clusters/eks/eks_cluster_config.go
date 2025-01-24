package eks

import (
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
)

const (
	// The json/yaml config key for the EKS hosted cluster config
	EKSClusterConfigConfigurationFileKey = "eksClusterConfig"
)

// ClusterConfig is the configuration needed to create an EKS host cluster
type ClusterConfig struct {
	EBSCSIDriver        *bool              `json:"ebsCSIDriver,omitempty" yaml:"ebsCSIDriver,omitempty"`
	KmsKey              *string            `json:"kmsKey,omitempty" yaml:"kmsKey,omitempty"`
	KubernetesVersion   *string            `json:"kubernetesVersion,omitempty" yaml:"kubernetesVersion,omitempty"`
	LoggingTypes        []string           `json:"loggingTypes" yaml:"loggingTypes"`
	NodeGroupsConfig    *[]NodeGroupConfig `json:"nodeGroups,omitempty" yaml:"nodeGroups,omitempty"`
	PrivateAccess       *bool              `json:"privateAccess,omitempty" yaml:"privateAccess,omitempty"`
	PublicAccess        *bool              `json:"publicAccess,omitempty" yaml:"publicAccess,omitempty"`
	PublicAccessSources []string           `json:"publicAccessSources" yaml:"publicAccessSources"`
	Region              string             `json:"region" yaml:"region"`
	SecretsEncryption   *bool              `json:"secretsEncryption,omitempty" yaml:"secretsEncryption,omitempty"`
	SecurityGroups      []string           `json:"securityGroups" yaml:"securityGroups"`
	ServiceRole         *string            `json:"serviceRole,omitempty" yaml:"serviceRole,omitempty"`
	Subnets             []string           `json:"subnets" yaml:"subnets"`
	Tags                map[string]string  `json:"tags" yaml:"tags"`
}

// NodeGroupConfig is the configuration need to create an EKS node group
type NodeGroupConfig struct {
	Arm                  *bool                 `json:"arm,omitempty" yaml:"arm,omitempty"`
	DesiredSize          *int64                `json:"desiredSize,omitempty" yaml:"desiredSize,omitempty"`
	DiskSize             *int64                `json:"diskSize,omitempty" yaml:"diskSize,omitempty"`
	Ec2SshKey            *string               `json:"ec2SshKey,omitempty" yaml:"ec2SshKey,omitempty"`
	Gpu                  *bool                 `json:"gpu,omitempty" yaml:"gpu,omitempty"`
	ImageID              *string               `json:"imageId,omitempty" yaml:"imageId,omitempty"`
	InstanceType         *string               `json:"instanceType,omitempty" yaml:"instanceType,omitempty"`
	Labels               map[string]string     `json:"labels" yaml:"labels"`
	LaunchTemplateConfig *LaunchTemplateConfig `json:"launchTemplate,omitempty" yaml:"launchTemplate,omitempty"`
	MaxSize              *int64                `json:"maxSize,omitempty" yaml:"maxSize,omitempty"`
	MinSize              *int64                `json:"minSize,omitempty" yaml:"minSize,omitempty"`
	NodeRole             *string               `json:"nodeRole,omitempty" yaml:"nodeRole,omitempty"`
	NodegroupName        *string               `json:"nodegroupName,omitempty" yaml:"nodegroupName,omitempty"`
	RequestSpotInstances *bool                 `json:"requestSpotInstances,omitempty" yaml:"requestSpotInstances,omitempty"`
	ResourceTags         map[string]string     `json:"resourceTags" yaml:"resourceTags"`
	SpotInstanceTypes    []string              `json:"spotInstanceTypes" yaml:"spotInstanceTypes"`
	Subnets              []string              `json:"subnets" yaml:"subnets"`
	Tags                 map[string]string     `json:"tags" yaml:"tags"`
	UserData             *string               `json:"userData,omitempty" yaml:"userData,omitempty"`
	Version              *string               `json:"version,omitempty" yaml:"version,omitempty"`
}

// LaunchTemplateConfig is the configuration need for a node group launch template
type LaunchTemplateConfig struct {
	ID      *string `json:"id,omitempty" yaml:"id,omitempty"`
	Name    *string `json:"name,omitempty" yaml:"name,omitempty"`
	Version *int64  `json:"version,omitempty" yaml:"version,omitempty"`
}

func nodeGroupsConstructor(nodeGroupsConfig *[]NodeGroupConfig, kubernetesVersion string) *[]management.NodeGroup {
	var nodeGroups = make([]management.NodeGroup, 0)
	if nodeGroupsConfig == nil {
		return nil
	}
	for _, nodeGroupConfig := range *nodeGroupsConfig {
		var launchTemplate *management.LaunchTemplate
		if nodeGroupConfig.LaunchTemplateConfig != nil {
			launchTemplate = &management.LaunchTemplate{
				ID:      nodeGroupConfig.LaunchTemplateConfig.ID,
				Name:    nodeGroupConfig.LaunchTemplateConfig.Name,
				Version: nodeGroupConfig.LaunchTemplateConfig.Version,
			}
		}
		var version *string
		if nodeGroupConfig.Version != nil {
			version = nodeGroupConfig.Version
		} else {
			version = &kubernetesVersion
		}
		nodeGroup := management.NodeGroup{
			Arm:                  nodeGroupConfig.Arm,
			DesiredSize:          nodeGroupConfig.DesiredSize,
			DiskSize:             nodeGroupConfig.DiskSize,
			Ec2SshKey:            nodeGroupConfig.Ec2SshKey,
			Gpu:                  nodeGroupConfig.Gpu,
			ImageID:              nodeGroupConfig.ImageID,
			InstanceType:         nodeGroupConfig.InstanceType,
			Labels:               &nodeGroupConfig.Labels,
			LaunchTemplate:       launchTemplate,
			MaxSize:              nodeGroupConfig.MaxSize,
			MinSize:              nodeGroupConfig.MinSize,
			NodegroupName:        nodeGroupConfig.NodegroupName,
			NodeRole:             nodeGroupConfig.NodeRole,
			RequestSpotInstances: nodeGroupConfig.RequestSpotInstances,
			ResourceTags:         &nodeGroupConfig.ResourceTags,
			SpotInstanceTypes:    &nodeGroupConfig.SpotInstanceTypes,
			Subnets:              &nodeGroupConfig.Subnets,
			Tags:                 &nodeGroupConfig.Tags,
			UserData:             nodeGroupConfig.UserData,
			Version:              version,
		}
		nodeGroups = append(nodeGroups, nodeGroup)
	}
	return &nodeGroups
}

func eksHostClusterConfig(displayName, cloudCredentialID string, eksClusterConfig ClusterConfig) *management.EKSClusterConfigSpec {
	return &management.EKSClusterConfigSpec{
		AmazonCredentialSecret: cloudCredentialID,
		DisplayName:            displayName,
		EBSCSIDriver:           eksClusterConfig.EBSCSIDriver,
		Imported:               false,
		KmsKey:                 eksClusterConfig.KmsKey,
		KubernetesVersion:      eksClusterConfig.KubernetesVersion,
		LoggingTypes:           &eksClusterConfig.LoggingTypes,
		NodeGroups:             nodeGroupsConstructor(eksClusterConfig.NodeGroupsConfig, *eksClusterConfig.KubernetesVersion),
		PrivateAccess:          eksClusterConfig.PrivateAccess,
		PublicAccess:           eksClusterConfig.PublicAccess,
		PublicAccessSources:    &eksClusterConfig.PublicAccessSources,
		Region:                 eksClusterConfig.Region,
		SecretsEncryption:      eksClusterConfig.SecretsEncryption,
		SecurityGroups:         &eksClusterConfig.SecurityGroups,
		ServiceRole:            eksClusterConfig.ServiceRole,
		Subnets:                &eksClusterConfig.Subnets,
		Tags:                   &eksClusterConfig.Tags,
	}
}
