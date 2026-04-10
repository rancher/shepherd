package features

import (
	"github.com/rancher/shepherd/clients/rancher"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IsFeatureEnabled is a helper function that uses wrangler context to check if a feature is enabled
func IsFeatureEnabled(client *rancher.Client, featureFlag string) (bool, error) {
	feature, err := client.WranglerContext.Mgmt.Feature().Get(featureFlag, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	if feature.Spec.Value == nil {
		return false, nil
	}

	return *feature.Spec.Value, nil
}

// UpdateFeatureFlag is a helper function that uses wrangler context to update a feature flag
func UpdateFeatureFlag(client *rancher.Client, featureFlag string, value bool) error {
	feature, err := client.WranglerContext.Mgmt.Feature().Get(featureFlag, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if feature.Spec.Value != nil && *feature.Spec.Value == value {
		return nil
	}

	feature.Spec.Value = &value
	_, err = client.WranglerContext.Mgmt.Feature().Update(feature)
	if err != nil {
		return err
	}

	return nil
}
