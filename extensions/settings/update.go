package settings

import (
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	v1 "github.com/rancher/shepherd/clients/rancher/v1"
)

// ResetUGlobalSettings is a helper function that uses the steve client to reset a specific
// Global Setting to its default value.
func ResetUGlobalSettings(steveclient *v1.Client, globalSetting *v1.SteveAPIObject) (*v1.SteveAPIObject, error) {
	updateSetting := &v3.Setting{}
	err := v1.ConvertToK8sType(globalSetting.JSONResp, updateSetting)
	if err != nil {
		return nil, err
	}

	updateSetting.Value = updateSetting.Default
	updateGlobalSetting, err := steveclient.SteveType(ManagementSetting).Update(globalSetting, updateSetting)
	if err != nil {
		return nil, err
	}
	return updateGlobalSetting, nil
}

// UpdateGlobalSettings is a helper function that uses the steve client to update a Global setting.
func UpdateGlobalSettings(steveclient *v1.Client, globalSetting *v1.SteveAPIObject, value string) (*v1.SteveAPIObject, error) {
	updateSetting := &v3.Setting{}
	err := v1.ConvertToK8sType(globalSetting.JSONResp, updateSetting)
	if err != nil {
		return nil, err
	}

	updateSetting.Value = value
	updateGlobalSetting, err := steveclient.SteveType(ManagementSetting).Update(globalSetting, updateSetting)
	if err != nil {
		return nil, err
	}
	return updateGlobalSetting, nil
}
