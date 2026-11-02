package aws

import (
	"fmt"

	"github.com/rancher/shepherd/extensions/cloudcredentials"
	"github.com/rancher/shepherd/pkg/config/operations"
)

// LoadAWSCredentialFromConfig is a helper function that loads AWS cloud credentials from the cattle config.
func LoadAWSCredentialFromConfig(cattleConfig map[string]any) (cloudcredentials.CloudCredential, error) {
	var credentialConfig cloudcredentials.AmazonEC2CredentialConfig
	if _, exists := cattleConfig[cloudcredentials.AmazonEC2CredentialConfigurationFileKey]; !exists {
		return cloudcredentials.CloudCredential{}, fmt.Errorf("missing credential key %q", cloudcredentials.AmazonEC2CredentialConfigurationFileKey)
	}

	operations.LoadObjectFromMap(cloudcredentials.AmazonEC2CredentialConfigurationFileKey, cattleConfig, &credentialConfig)

	return cloudcredentials.CloudCredential{AmazonEC2CredentialConfig: &credentialConfig}, nil
}
