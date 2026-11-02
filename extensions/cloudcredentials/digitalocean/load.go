package digitalocean

import (
	"fmt"

	"github.com/rancher/shepherd/extensions/cloudcredentials"
	"github.com/rancher/shepherd/pkg/config/operations"
)

// LoadDigitalOceanCredentialFromConfig is a helper function that loads DigitalOcean cloud credentials from the cattle config.
func LoadDigitalOceanCredentialFromConfig(cattleConfig map[string]any) (cloudcredentials.CloudCredential, error) {
	var credentialConfig cloudcredentials.DigitalOceanCredentialConfig
	if _, exists := cattleConfig[cloudcredentials.DigitalOceanCredentialConfigurationFileKey]; !exists {
		return cloudcredentials.CloudCredential{}, fmt.Errorf("missing credential key %q in cattleConfig", cloudcredentials.DigitalOceanCredentialConfigurationFileKey)
	}

	operations.LoadObjectFromMap(cloudcredentials.DigitalOceanCredentialConfigurationFileKey, cattleConfig, &credentialConfig)

	return cloudcredentials.CloudCredential{DigitalOceanCredentialConfig: &credentialConfig}, nil
}
