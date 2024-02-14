package nodes

import (
	"github.com/rancher/norman/types"
	v3 "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/kubeapi/secrets"
	corev1 "k8s.io/api/core/v1"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	initNodeLabelKey         = "rke.cattle.io/init-node"
	initNodeLabelValue       = "true"
	local                    = "local"
	machinePlanSecretType    = "rke.cattle.io/machine-plan"
	machineNameSteveLabel    = "rke.cattle.io/machine-name"
	node                     = "node"
)

// GetInitNode accepts a client and clusterID and returns the init node object
func GetInitNode(client *rancher.Client, clusterID string) (*v3.Node, error) {
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
		for key, val := range machinePlan.ObjectMeta.Labels {
			if key == initNodeLabelKey && val == initNodeLabelValue {
				logrus.Infof("Found init node machine plan: %s", machinePlan.Name)
				initNodeMachineName = machinePlan.Labels[machineNameSteveLabel]
			}
		}
	}

	logrus.Info("Retrieving list of nodes...")
	nodes, err := client.Management.Node.ListAll(&types.ListOpts{})
	if err != nil {
		return nil, err
	}

	initNodeID := ""

	logrus.Info("Identifying init node...")
	for _, node := range nodes.Data {
		for key, val := range node.Annotations {
			if key == machineSteveAnnotation && val == initNodeMachineName {
				logrus.Infof("Found init node: %s", node.ID)
				initNodeID = node.ID
			}
		}
	}

	logrus.Info("Retrieving init node...")
	initNode, err := client.Management.Node.ByID(initNodeID)
	if err != nil {
		return nil, err
	}

	return initNode, nil
}

// DeleteNodeRKE2K3S accepts a client and node object and deletes the node from the cluster
func DeleteNodeRKE2K3S(client *rancher.Client, node *v3.Node) error {
	logrus.Info("Targeting node for deletion...")
	machine, err := client.Steve.SteveType(machineSteveResourceType).ByID(fleetNamespace + "/" + node.Annotations[machineSteveAnnotation])
	if err != nil {
		return  err
	}

	logrus.Info("Deleting node...")
	err = client.Steve.SteveType(machineSteveResourceType).Delete(machine)
	if err != nil {
		return  err
	}

	return nil
}

// VerifyDeletedNodeRKE2K3S() accepts a client and node object and verifies that the node was successfully removed from the cluster
func VerifyDeletedNodeRKE2K3S(client *rancher.Client, node *v3.Node) error {
	logrus.Info("Re-fetching nodes...")

	nodes, err := client.Management.Node.ListAll(&types.ListOpts{})
	if err != nil {
		return err
	}

	logrus.Info("Verifying node was successfully removed...")
	for _, machine := range nodes.Data {
		if machine.ID == node.ID {
			logrus.Fatal("Node was not successfully removed from cluster")
		}
	}

	logrus.Info("Node was successfully removed from cluster!")
	return nil
}