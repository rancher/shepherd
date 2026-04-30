package features

import (
	"github.com/rancher/shepherd/clients/rancher"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IsFeatureEnabled returns true when the named feature flag is enabled.
func IsFeatureEnabled(client *rancher.Client, name string) (bool, error) {
	feature, err := client.WranglerContext.Mgmt.Feature().Get(name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	if feature.Spec.Value == nil {
		return false, nil
	}
	return *feature.Spec.Value, nil
}

// UpdateFeatureFlag sets the named feature flag to value; no-op if already at value.
func UpdateFeatureFlag(client *rancher.Client, name string, value bool) error {
	feature, err := client.WranglerContext.Mgmt.Feature().Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if feature.Spec.Value != nil && *feature.Spec.Value == value {
		return nil
	}
	feature.Spec.Value = &value
	_, err = client.WranglerContext.Mgmt.Feature().Update(feature)
	return err
}

// EnableFeatureFlag enables the named feature flag and registers DisableFeatureFlag as session cleanup.
func EnableFeatureFlag(client *rancher.Client, name string) error {
	enabled, err := IsFeatureEnabled(client, name)
	if err != nil {
		return err
	}
	if enabled {
		return nil
	}
	client.Session.RegisterCleanupFunc(func() error {
		return DisableFeatureFlag(client, name)
	})
	return UpdateFeatureFlag(client, name, true)
}

// DisableFeatureFlag disables the named feature flag.
func DisableFeatureFlag(client *rancher.Client, name string) error {
	return UpdateFeatureFlag(client, name, false)
}
