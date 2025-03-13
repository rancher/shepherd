package vsphere

import (
	"github.com/rancher/shepherd/clients/rancher"
	v1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/extensions/cloudcredentials"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	"github.com/rancher/shepherd/extensions/defaults/stevestates"
	"github.com/rancher/shepherd/extensions/defaults/stevetypes"
	"github.com/rancher/shepherd/extensions/steve"
	"github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/namegenerator"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	vsphereProvider = "vsphere"
	localCluster    = "local"
)

// CreateVsphereCloudCredentials is a helper function that creates V1 cloud credentials and waits for them to become active.
func CreateVsphereCloudCredentials(client *rancher.Client, credentials cloudcredentials.CloudCredential) (*v1.SteveAPIObject, error) {
	secretName := namegenerator.AppendRandomString(vsphereProvider)
	spec := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: cloudcredentials.GeneratedName,
			Namespace:    namespaces.CattleData,
			Annotations: map[string]string{
				"field.cattle.io/name":      secretName,
				"field.cattle.io/creatorId": client.UserID,
			},
		},
		Data: map[string][]byte{
			"vmwarevspherecredentialConfig-password":    []byte(credentials.VmwareVsphereConfig.Password),
			"vmwarevspherecredentialConfig-username":    []byte(credentials.VmwareVsphereConfig.Username),
			"vmwarevspherecredentialConfig-vcenter":     []byte(credentials.VmwareVsphereConfig.Vcenter),
			"vmwarevspherecredentialConfig-vcenterPort": []byte(credentials.VmwareVsphereConfig.VcenterPort),
		},
		Type: corev1.SecretTypeOpaque,
	}

	vSphereCloudCredentials, err := steve.CreateAndWaitForResource(client, namespaces.FleetLocal+"/"+localCluster, stevetypes.Secret, spec, stevestates.Active, defaults.FiveSecondTimeout, defaults.FiveMinuteTimeout)
	if err != nil {
		return nil, err
	}

	return vSphereCloudCredentials, nil
}

// GetVspherePassword is a helper to get the password from the cloud credential object as a string
func GetVspherePassword() string {
	var vmwarevsphereCredentialConfig cloudcredentials.VmwarevsphereCredentialConfig

	config.LoadConfig(cloudcredentials.VmwarevsphereCredentialConfigurationFileKey, &vmwarevsphereCredentialConfig)

	return vmwarevsphereCredentialConfig.Password
}
