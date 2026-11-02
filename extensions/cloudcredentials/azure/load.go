package azure

import (
	"fmt"

	"github.com/rancher/shepherd/extensions/cloudcredentials"
	"github.com/rancher/shepherd/pkg/config/operations"
)

// LoadAzureCredentialFromConfig is a helper function that loads Azure cloud credentials from the cattle config.
func LoadAzureCredentialFromConfig(cattleConfig map[string]any) (cloudcredentials.CloudCredential, error) {
	var credentialConfig cloudcredentials.AzureCredentialConfig
	if _, exists := cattleConfig[cloudcredentials.AzureCredentialConfigurationFileKey]; !exists {
		return cloudcredentials.CloudCredential{}, fmt.Errorf("missing credential key %q in cattleConfig", cloudcredentials.AzureCredentialConfigurationFileKey)
	}

	operations.LoadObjectFromMap(cloudcredentials.AzureCredentialConfigurationFileKey, cattleConfig, &credentialConfig)

	return cloudcredentials.CloudCredential{AzureCredentialConfig: &credentialConfig}, nil
}
