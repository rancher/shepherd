package cloudcredentials

import (
	"fmt"

	"github.com/rancher/norman/types"
	"github.com/rancher/shepherd/extensions/defaults/providers"
	"github.com/rancher/shepherd/pkg/config"
)

const (
	GeneratedName = "cc-"
)

// CloudCredential is the main struct needed to create a cloud credential depending on the outside cloud service provider
type CloudCredential struct {
	types.Resource
	Annotations                  map[string]string              `json:"annotations,omitempty"`
	Created                      string                         `json:"created,omitempty"`
	CreatorID                    string                         `json:"creatorId,omitempty"`
	Description                  string                         `json:"description,omitempty"`
	Labels                       map[string]string              `json:"labels,omitempty"`
	Name                         string                         `json:"name,omitempty"`
	Removed                      string                         `json:"removed,omitempty"`
	AmazonEC2CredentialConfig    *AmazonEC2CredentialConfig     `json:"amazonec2credentialConfig,omitempty"`
	AzureCredentialConfig        *AzureCredentialConfig         `json:"azurecredentialConfig,omitempty"`
	DigitalOceanCredentialConfig *DigitalOceanCredentialConfig  `json:"digitaloceancredentialConfig,omitempty"`
	LinodeCredentialConfig       *LinodeCredentialConfig        `json:"linodecredentialConfig,omitempty"`
	HarvesterCredentialConfig    *HarvesterCredentialConfig     `json:"harvestercredentialConfig,omitempty"`
	GoogleCredentialConfig       *GoogleCredentialConfig        `json:"googlecredentialConfig,omitempty"`
	VmwareVsphereConfig          *VmwarevsphereCredentialConfig `json:"vmwarevspherecredentialConfig,omitempty"`
	AlibabaCredentialConfig      *AlibabaCredentialConfig       `json:"alibabacredentialConfig,omitempty"`
	UUID                         string                         `json:"uuid,omitempty"`
}

// LoadCloudCredential loads the providers cloudCredentialConfig from the cattle config file
func LoadCloudCredential(provider string) CloudCredential {
	var cloudCredential CloudCredential
	switch {

	case provider == providers.AWS:
		var awsCredentialConfig AmazonEC2CredentialConfig

		config.LoadConfig(AmazonEC2CredentialConfigurationFileKey, &awsCredentialConfig)
		cloudCredential.AmazonEC2CredentialConfig = &awsCredentialConfig

		return cloudCredential

	case provider == providers.Azure:
		var azureCredentialConfig AzureCredentialConfig

		config.LoadConfig(AzureCredentialConfigurationFileKey, &azureCredentialConfig)
		cloudCredential.AzureCredentialConfig = &azureCredentialConfig

		return cloudCredential

	case provider == providers.DigitalOcean:
		var digitalOceanCredentialConfig DigitalOceanCredentialConfig

		config.LoadConfig(DigitalOceanCredentialConfigurationFileKey, &digitalOceanCredentialConfig)
		cloudCredential.DigitalOceanCredentialConfig = &digitalOceanCredentialConfig

		return cloudCredential

	case provider == providers.Linode:
		var linodeCredentialConfig LinodeCredentialConfig

		config.LoadConfig(LinodeCredentialConfigurationFileKey, &linodeCredentialConfig)
		cloudCredential.LinodeCredentialConfig = &linodeCredentialConfig

		return cloudCredential

	case provider == providers.Harvester:
		var harvesterCredentialConfig HarvesterCredentialConfig

		config.LoadConfig(HarvesterCredentialConfigurationFileKey, &harvesterCredentialConfig)
		cloudCredential.HarvesterCredentialConfig = &harvesterCredentialConfig

		return cloudCredential

	case provider == providers.Vsphere:
		var vsphereCredentialConfig VmwarevsphereCredentialConfig

		config.LoadConfig(VmwarevsphereCredentialConfigurationFileKey, &vsphereCredentialConfig)
		cloudCredential.VmwareVsphereConfig = &vsphereCredentialConfig

		return cloudCredential

	case provider == providers.Google:
		var googleCredentialConfig GoogleCredentialConfig

		config.LoadConfig(GoogleCredentialConfigurationFileKey, &googleCredentialConfig)
		cloudCredential.GoogleCredentialConfig = &googleCredentialConfig

		return cloudCredential

	case provider == providers.Alibaba:
		var alibabaCredentialConfig AlibabaCredentialConfig

		config.LoadConfig(AlibabaCredentialConfigurationFileKey, &alibabaCredentialConfig)
		cloudCredential.AlibabaCredentialConfig = &alibabaCredentialConfig

		return cloudCredential

	default:
		panic(fmt.Sprintf("Provider:%v not found", provider))
	}
}
