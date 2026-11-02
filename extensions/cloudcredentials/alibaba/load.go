package alibaba

import (
	"fmt"

	"github.com/rancher/shepherd/extensions/cloudcredentials"
	"github.com/rancher/shepherd/pkg/config/operations"
)

// LoadAlibabaCredentialFromConfig is a helper function that loads Alibaba cloud credentials from the cattle config.
func LoadAlibabaCredentialFromConfig(cattleConfig map[string]any) (cloudcredentials.CloudCredential, error) {
	var credentialConfig cloudcredentials.AlibabaCredentialConfig
	if _, exists := cattleConfig[cloudcredentials.AlibabaCredentialConfigurationFileKey]; !exists {
		return cloudcredentials.CloudCredential{}, fmt.Errorf("missing credential key %q in cattleConfig", cloudcredentials.AlibabaCredentialConfigurationFileKey)
	}

	operations.LoadObjectFromMap(cloudcredentials.AlibabaCredentialConfigurationFileKey, cattleConfig, &credentialConfig)

	return cloudcredentials.CloudCredential{AlibabaCredentialConfig: &credentialConfig}, nil
}
