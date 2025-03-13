package digitalocean

import (
	"github.com/rancher/shepherd/clients/rancher"
	v1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/extensions/cloudcredentials"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	"github.com/rancher/shepherd/extensions/defaults/providers"
	"github.com/rancher/shepherd/extensions/defaults/stevestates"
	"github.com/rancher/shepherd/extensions/defaults/stevetypes"
	"github.com/rancher/shepherd/extensions/steve"
	"github.com/rancher/shepherd/pkg/namegenerator"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	localCluster = "local"
)

// CreateDigitalOceanCloudCredentials is a helper function that creates V1 cloud credentials and waits for them to become active.
func CreateDigitalOceanCloudCredentials(client *rancher.Client, credentials cloudcredentials.CloudCredential) (*v1.SteveAPIObject, error) {
	secretName := namegenerator.AppendRandomString(providers.DigitalOcean)
	spec := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: cloudcredentials.GeneratedName,
			Namespace:    namespaces.CattleData,
			Annotations: map[string]string{
				"provisioning.cattle.io/driver": providers.DigitalOcean,
				"field.cattle.io/name":          secretName,
				"field.cattle.io/creatorId":     client.UserID,
			},
		},
		Data: map[string][]byte{
			"digitaloceancredentialConfig-accessToken": []byte(credentials.DigitalOceanCredentialConfig.AccessToken),
		},
		Type: corev1.SecretTypeOpaque,
	}

	digitalOceanCloudCredentials, err := steve.CreateAndWaitForResource(client, namespaces.FleetLocal+"/"+localCluster, stevetypes.Secret, spec, stevestates.Active, defaults.FiveSecondTimeout, defaults.FiveMinuteTimeout)
	if err != nil {
		return nil, err
	}

	return digitalOceanCloudCredentials, nil
}
