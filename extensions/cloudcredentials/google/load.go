package google

import (
	"fmt"

	"github.com/rancher/shepherd/extensions/cloudcredentials"
	"github.com/rancher/shepherd/pkg/config/operations"
)

// LoadGoogleCredentialFromConfig is a helper function that loads Google cloud credentials from the cattle config.
func LoadGoogleCredentialFromConfig(cattleConfig map[string]any) (cloudcredentials.CloudCredential, error) {
	var credentialConfig cloudcredentials.GoogleCredentialConfig
	if _, exists := cattleConfig[cloudcredentials.GoogleCredentialConfigurationFileKey]; !exists {
		return cloudcredentials.CloudCredential{}, fmt.Errorf("missing credential key %q in cattleConfig", cloudcredentials.GoogleCredentialConfigurationFileKey)
	}

	operations.LoadObjectFromMap(cloudcredentials.GoogleCredentialConfigurationFileKey, cattleConfig, &credentialConfig)

	return cloudcredentials.CloudCredential{GoogleCredentialConfig: &credentialConfig}, nil
}
