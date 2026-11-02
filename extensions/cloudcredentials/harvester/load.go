package harvester

import (
	"fmt"

	"github.com/rancher/shepherd/extensions/cloudcredentials"
	"github.com/rancher/shepherd/pkg/config/operations"
)

// LoadHarvesterCredentialFromConfig is a helper function that loads Harvester cloud credentials from the cattle config.
func LoadHarvesterCredentialFromConfig(cattleConfig map[string]any) (cloudcredentials.CloudCredential, error) {
	var credentialConfig cloudcredentials.HarvesterCredentialConfig
	if _, exists := cattleConfig[cloudcredentials.HarvesterCredentialConfigurationFileKey]; !exists {
		return cloudcredentials.CloudCredential{}, fmt.Errorf("missing credential key %q in cattleConfig", cloudcredentials.HarvesterCredentialConfigurationFileKey)
	}

	operations.LoadObjectFromMap(cloudcredentials.HarvesterCredentialConfigurationFileKey, cattleConfig, &credentialConfig)

	return cloudcredentials.CloudCredential{HarvesterCredentialConfig: &credentialConfig}, nil
}
