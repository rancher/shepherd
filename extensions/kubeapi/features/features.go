package features

import (
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/kubeapi/cluster"
	"github.com/rancher/shepherd/extensions/kubeapi/workloads/deployments"
	"github.com/sirupsen/logrus"
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

// UpdateFeatureFlag updates the value of the named feature flag to the provided value using the wrangler context.
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

// EnableFeatureFlag enables the named feature flag and registers session cleanup that disables it again
// and waits for the Rancher rollout to settle, so cleanup does not hand back a restarting server. The flag
// is left untouched, and no cleanup registered, when it is already enabled.
func EnableFeatureFlag(client *rancher.Client, name string) error {
	enabled, err := IsFeatureEnabled(client, name)
	if err != nil {
		return err
	}

	if enabled {
		return nil
	}

	client.Session.RegisterCleanupFunc(func() error {
		if err := DisableFeatureFlag(client, name); err != nil {
			return err
		}

		// The session retries a cleanup func that returns an error, and re-waiting a rollout that already
		// timed out only multiplies the delay, so the wait is reported rather than returned.
		if err := deployments.WaitForDeploymentActive(client, cluster.LocalCluster, deployments.RancherDeploymentNamespace, deployments.RancherDeploymentName); err != nil {
			logrus.Errorf("rancher did not settle after disabling feature flag %s: %v", name, err)
		}

		return nil
	})

	return UpdateFeatureFlag(client, name, true)
}

// DisableFeatureFlag disables the named feature flag using the wrangler context.
func DisableFeatureFlag(client *rancher.Client, name string) error {
	return UpdateFeatureFlag(client, name, false)
}
