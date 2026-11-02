package linode

import (
	"fmt"

	"github.com/rancher/shepherd/extensions/cloudcredentials"
	"github.com/rancher/shepherd/pkg/config/operations"
)

// LoadLinodeCredentialFromConfig is a helper function that loads Linode cloud credentials from the cattle config.
func LoadLinodeCredentialFromConfig(cattleConfig map[string]any) (cloudcredentials.CloudCredential, error) {
	var credentialConfig cloudcredentials.LinodeCredentialConfig
	if _, exists := cattleConfig[cloudcredentials.LinodeCredentialConfigurationFileKey]; !exists {
		return cloudcredentials.CloudCredential{}, fmt.Errorf("missing credential key %q in cattleConfig", cloudcredentials.LinodeCredentialConfigurationFileKey)
	}

	operations.LoadObjectFromMap(cloudcredentials.LinodeCredentialConfigurationFileKey, cattleConfig, &credentialConfig)

	return cloudcredentials.CloudCredential{LinodeCredentialConfig: &credentialConfig}, nil
}
