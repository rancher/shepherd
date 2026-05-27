package settings

import (
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AutoscalerChartRepo                  = "cluster-autoscaler-chart-repository"
	AutoscalerImage                      = "cluster-autoscaler-image"
	AuthUserSessionIdleTTlMinutesSetting = "auth-user-session-idle-ttl-minutes"
	AuthTokenMaxTTLMinutesSetting        = "auth-token-max-ttl-minutes"
	AuthTokenMaxTTLMinutes               = "auth-token-max-ttl-minutes"
	KubeconfigDefaultTTLMinutes          = "kubeconfig-default-token-ttl-minutes"
	UserPasswordMinLength                = "password-min-length"
)

// GetGlobalSettingNames is a helper function that uses wrangler context to fetch a list of global setting names
func GetGlobalSettingNames(client *rancher.Client) ([]string, error) {
	settings, err := client.WranglerContext.Mgmt.Setting().List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list settings: %w", err)
	}

	globalSettings := []string{}
	for _, gs := range settings.Items {
		globalSettings = append(globalSettings, gs.Name)
	}

	return globalSettings, nil
}

// UpdateGlobalSetting is a helper function that uses the wrangler context to update the value of a Rancher global setting given its ID
func UpdateGlobalSetting(client *rancher.Client, settingID, value string) error {
	setting, err := client.WranglerContext.Mgmt.Setting().Get(settingID, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get setting %s: %w", settingID, err)
	}

	setting.Value = value

	updatedSetting, err := client.WranglerContext.Mgmt.Setting().Update(setting)
	if err != nil {
		return fmt.Errorf("failed to update setting %s: %w", settingID, err)
	}

	if updatedSetting.Value != value {
		return fmt.Errorf("failed to update setting %q; got: %s, expected: %s",
			settingID, updatedSetting.Value, value)
	}

	currentSetting, err := client.WranglerContext.Mgmt.Setting().Get(settingID, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to verify setting %s after update: %w", settingID, err)
	}

	if currentSetting.Value != value {
		return fmt.Errorf("setting %q was not persisted; got: %s, expected: %s",
			settingID, currentSetting.Value, value)
	}

	return nil
}

// ResetGlobalSettingToDefaultValue is a helper function that uses the wrangler context to reset a global setting by name to it's default value
func ResetGlobalSettingToDefaultValue(client *rancher.Client, settingName string) error {
	defaultValue, err := GetGlobalSettingDefaultValue(client, settingName)
	if err != nil {
		return fmt.Errorf("failed to get default value for setting %s: %w", settingName, err)
	}

	return UpdateGlobalSetting(client, settingName, defaultValue)
}

// GetGlobalSettingDefaultValue is a helper function that uses the wrangler context to retrieve the default value of a Rancher global setting given its ID
func GetGlobalSettingDefaultValue(client *rancher.Client, settingName string) (string, error) {
	setting, err := client.WranglerContext.Mgmt.Setting().Get(settingName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get setting %s: %w", settingName, err)
	}

	return setting.Default, nil
}
