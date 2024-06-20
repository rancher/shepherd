package features

import (
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	v1 "github.com/rancher/shepherd/clients/rancher/v1"
)

// UpdateFeatureFlag is a helper function that uses the steve client to update a Global setting
func UpdateFeatureFlag(steveclient *v1.Client, featureFlag *v1.SteveAPIObject, value bool) (*v1.SteveAPIObject, error) {
	updateFeature := &v3.Feature{}
	err := v1.ConvertToK8sType(featureFlag.JSONResp, updateFeature)
	if err != nil {
		return nil, err
	}

	updateFeature.Spec.Value = &value

	updateFeatureFlag, err := steveclient.SteveType(ManagementFeature).Update(featureFlag, updateFeature)
	if err != nil {
		return nil, err
	}

	return updateFeatureFlag, nil
}
