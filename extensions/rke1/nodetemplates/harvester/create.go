package nodetemplates

import (
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/cloudcredentials"
	"github.com/rancher/shepherd/extensions/cloudcredentials/harvester"
	"github.com/rancher/shepherd/extensions/rke1/nodetemplates"
	"github.com/rancher/shepherd/pkg/config"
)

const harvesterNodeTemplateNameBase = "harvesterNodeConfig"

// CreateHarvesterNodeTemplate is a helper function that takes the rancher Client as a parameter and creates
// an Harvester node template and returns the NodeTemplate response
func CreateHarvesterNodeTemplate(rancherClient *rancher.Client) (*nodetemplates.NodeTemplate, error) {
	var harvesterNodeTemplateConfig nodetemplates.HarvesterNodeTemplateConfig
	config.LoadConfig(nodetemplates.HarvesterNodeTemplateConfigurationFileKey, &harvesterNodeTemplateConfig)

	var cloudCredentialConfig cloudcredentials.CloudCredential
	config.LoadConfig(cloudcredentials.HarvesterCredentialConfigurationFileKey, &cloudCredentialConfig.HarvesterCredentialConfig)
	cloudCredential, err := harvester.CreateHarvesterCloudCredentials(rancherClient, cloudCredentialConfig)
	if err != nil {
		return nil, err
	}

	nodeTemplate := nodetemplates.NodeTemplate{
		EngineInstallURL:            "https://releases.rancher.com/install-docker/24.0.sh",
		Name:                        harvesterNodeTemplateNameBase,
		HarvesterNodeTemplateConfig: &harvesterNodeTemplateConfig,
	}

	nodeTemplateConfig := &nodetemplates.NodeTemplate{
		CloudCredentialID: cloudCredential.Namespace + ":" + cloudCredential.Name,
	}

	config.LoadConfig(nodetemplates.NodeTemplateConfigurationFileKey, nodeTemplateConfig)

	nodeTemplateFinal, err := nodeTemplate.MergeOverride(nodeTemplateConfig, nodetemplates.HarvesterNodeTemplateConfigurationFileKey)
	if err != nil {
		return nil, err
	}

	resp := &nodetemplates.NodeTemplate{}
	err = rancherClient.Management.APIBaseClient.Ops.DoCreate(management.NodeTemplateType, *nodeTemplateFinal, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
