package azure

import (
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/cloudcredentials"
	"github.com/rancher/shepherd/pkg/config"
)

const azureCloudCredNameBase = "azureOceanCloudCredential"

// CreateAzureCloudCredentials is a helper function that takes the rancher Client as a parameter and creates
// an Azure cloud credential, and returns the CloudCredential response
func CreateAzureCloudCredentials(rancherClient *rancher.Client) (*cloudcredentials.CloudCredential, error) {
	var azureCredentialConfig cloudcredentials.AzureCredentialConfig
	config.LoadConfig(cloudcredentials.AzureCredentialConfigurationFileKey, &azureCredentialConfig)

	cloudCredential := cloudcredentials.CloudCredential{
		Name:                  azureCloudCredNameBase,
		AzureCredentialConfig: &azureCredentialConfig,
	}

	resp := &cloudcredentials.CloudCredential{}
	err := rancherClient.Management.APIBaseClient.Ops.DoCreate(management.CloudCredentialType, cloudCredential, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
