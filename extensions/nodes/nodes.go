package nodes

import (
	"errors"

	"github.com/rancher/shepherd/clients/rancher"
	v1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/extensions/kubeapi/secrets"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	initNodeLabelKey      = "rke.cattle.io/init-node"
	local                 = "local"
	machinePlanSecretType = "rke.cattle.io/machine-plan"
	machineNameSteveLabel = "rke.cattle.io/machine-name"
)

// GetInitMachine accepts a client and clusterID and returns the "init node" machine
// object for rke2/k3s clusters
func GetInitMachine(client *rancher.Client, clusterID string) (*v1.SteveAPIObject, error) {
	logrus.Info("Retrieving list of secrets...")
	secretList, err := secrets.ListSecrets(client, local, fleetNamespace, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	machinePlans := []corev1.Secret{}

	logrus.Info("Identifying machine-plan secrets...")
	for _, secret := range secretList.Items {
		if secret.Type == machinePlanSecretType {
			logrus.Infof("Found machine-plan secret: %s", secret.Name)
			machinePlans = append(machinePlans, secret)
		}
	}

	initNodeMachineName := ""

	logrus.Info("Identifying init node machine plan...")
	for _, machinePlan := range machinePlans {
		v, ok := machinePlan.ObjectMeta.Labels[initNodeLabelKey]
		if ok && v == "true" {
			logrus.Infof("Found init node machine plan: %s", machinePlan.Name)
			initNodeMachineName = machinePlan.Labels[machineNameSteveLabel]
		}
	}

	logrus.Info("Retrieving machine...")
	initMachine, err := client.Steve.SteveType(machineSteveResourceType).ByID(fleetNamespace + "/" + initNodeMachineName)
	if err != nil {
		return nil, err
	}

	logrus.Infof("Successfully retrieved machine: %s", initNodeMachineName)

	return initMachine, nil
}

// DeleteMachineRKE2K3S accepts a client and v1.SteveAPIObject and deletes the machine
func DeleteMachineRKE2K3S(client *rancher.Client, machine *v1.SteveAPIObject) error {
	logrus.Info("Deleting machine...")
	err := client.Steve.SteveType(machineSteveResourceType).Delete(machine)
	if err != nil {
		return err
	}

	return nil
}

// VerifyDeletedMachineRKE2K3S accepts a client and v1.SteveAPIObject and verifies that
// the machine was successfully removed from the cluster
func VerifyDeletedMachineRKE2K3S(client *rancher.Client, deletedMachine *v1.SteveAPIObject) error {
	logrus.Info("Re-fetching machines...")

	machines, err := client.Steve.SteveType(machineSteveResourceType).ListAll(nil)
	if err != nil {
		return err
	}

	logrus.Info("Verifying machine was successfully removed...")
	for _, machine := range machines.Data {
		if machine.Resource.ID == deletedMachine.Resource.ID {
			return errors.New("Machine was not successfully removed from cluster")
		}
	}

	logrus.Info("Machine was successfully removed from cluster!")
	return nil
}
