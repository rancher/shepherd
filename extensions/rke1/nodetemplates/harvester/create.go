package nodetemplates

import (
	"dario.cat/mergo"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/rke1/nodetemplates"
	"github.com/rancher/shepherd/pkg/config"
)

const harvesterNodeTemplateNameBase = "harvesterNodeConfig"

// CreateHarvesterNodeTemplate is a helper function that takes the rancher Client as a parameter and creates
// an Harvester node template and returns the NodeTemplate response
func CreateHarvesterNodeTemplate(rancherClient *rancher.Client) (*nodetemplates.NodeTemplate, error) {
	var harvesterNodeTemplateConfig nodetemplates.HarvesterNodeTemplateConfig
	config.LoadConfig(nodetemplates.HarvesterNodeTemplateConfigurationFileKey, &harvesterNodeTemplateConfig)

	nodeTemplate := nodetemplates.NodeTemplate{
		EngineInstallURL:            "https://releases.rancher.com/install-docker/24.0.sh",
		Name:                        harvesterNodeTemplateNameBase,
		HarvesterNodeTemplateConfig: &harvesterNodeTemplateConfig,
	}

	nodeTemplateConfig := &nodetemplates.NodeTemplate{}
	config.LoadConfig(nodetemplates.NodeTemplateConfigurationFileKey, nodeTemplateConfig)

	err := mergo.Merge(&nodeTemplate, nodeTemplateConfig, mergo.WithOverride)
	if err != nil {
		return nil, err
	}

	resp := &nodetemplates.NodeTemplate{}
	err = rancherClient.Management.APIBaseClient.Ops.DoCreate(management.NodeTemplateType, nodeTemplate, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
