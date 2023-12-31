package harvester

import (
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/cloudcredentials"
	"github.com/rancher/shepherd/pkg/config"
)

const harvesterCloudCredNameBase = "harvesterCloudCredential"

// CreateHarvesterCloudCredentials is a helper function that takes the rancher Client as a parameter and creates
// a harvester cloud credential, and returns the CloudCredential response
func CreateHarvesterCloudCredentials(rancherClient *rancher.Client) (*cloudcredentials.CloudCredential, error) {
	var harvesterCredentialConfig cloudcredentials.HarvesterCredentialConfig
	config.LoadConfig(cloudcredentials.HarvesterCredentialConfigurationFileKey, &harvesterCredentialConfig)

	cloudCredential := cloudcredentials.CloudCredential{
		Name:                      harvesterCloudCredNameBase,
		HarvesterCredentialConfig: &harvesterCredentialConfig,
	}

	resp := &cloudcredentials.CloudCredential{}
	err := rancherClient.Management.APIBaseClient.Ops.DoCreate(management.CloudCredentialType, cloudCredential, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
