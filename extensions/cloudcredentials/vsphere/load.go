package vsphere

import (
	"fmt"

	"github.com/rancher/shepherd/extensions/cloudcredentials"
	"github.com/rancher/shepherd/pkg/config/operations"
)

// LoadVSphereCredentialFromConfig is a helper function that loads vSphere cloud credentials from the cattle config.
func LoadVSphereCredentialFromConfig(cattleConfig map[string]any) (cloudcredentials.CloudCredential, error) {
	var credentialConfig cloudcredentials.VmwarevsphereCredentialConfig
	if _, exists := cattleConfig[cloudcredentials.VmwarevsphereCredentialConfigurationFileKey]; !exists {
		return cloudcredentials.CloudCredential{}, fmt.Errorf("missing credential key %q in cattleConfig", cloudcredentials.VmwarevsphereCredentialConfigurationFileKey)
	}

	operations.LoadObjectFromMap(cloudcredentials.VmwarevsphereCredentialConfigurationFileKey, cattleConfig, &credentialConfig)

	return cloudcredentials.CloudCredential{VmwareVsphereConfig: &credentialConfig}, nil
}
