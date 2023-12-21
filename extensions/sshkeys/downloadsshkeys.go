package sshkeys

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	provv1 "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	"github.com/rancher/shepherd/clients/rancher"
	steveV1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/extensions/clusters"
	kubeapinodes "github.com/rancher/shepherd/extensions/kubeapi/nodes"
	"github.com/rancher/shepherd/pkg/nodes"
	corev1 "k8s.io/api/core/v1"
)

const (
	privateKeySSHKeyRegExPattern              = `-----BEGIN RSA PRIVATE KEY-{3,}\n([\s\S]*?)\n-{3,}END RSA PRIVATE KEY-----`
	ClusterMachineConstraintResourceSteveType = "cluster.x-k8s.io.machine"
	ClusterMachineAnnotation                  = "cluster.x-k8s.io/machine"

	rootUser = "root"
)

// DownloadSSHKeys is a helper function that takes a client, the machinePoolNodeName to download
// the ssh key for a particular node.
func DownloadSSHKeys(client *rancher.Client, machinePoolNodeName string) ([]byte, error) {
	machinePoolNodeNameName := fmt.Sprintf("fleet-default/%s", machinePoolNodeName)
	machine, err := client.Steve.SteveType(ClusterMachineConstraintResourceSteveType).ByID(machinePoolNodeNameName)
	if err != nil {
		return nil, err
	}

	sshKeyLink := machine.Links["sshkeys"]

	req, err := http.NewRequest("GET", sshKeyLink, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+client.RancherConfig.AdminToken)

	resp, err := client.Management.APIBaseClient.Ops.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	privateSSHKeyRegEx := regexp.MustCompile(privateKeySSHKeyRegExPattern)
	privateSSHKey := privateSSHKeyRegEx.FindString(string(bodyBytes))

	return []byte(privateSSHKey), err
}

// GetSSHNodeFromMachine returns the v1/node object given a steve/v1/machine object.
func GetSSHNodeFromMachine(client *rancher.Client, sshUser string, machine *steveV1.SteveAPIObject) (*nodes.Node, error) {
	machineName := machine.Annotations[ClusterMachineAnnotation]
	sshkey, err := DownloadSSHKeys(client, machineName)
	if err != nil {
		return nil, err
	}

	newNode := &corev1.Node{}
	err = steveV1.ConvertToK8sType(machine.JSONResp, newNode)
	if err != nil {
		return nil, err
	}

	nodeIP := kubeapinodes.GetNodeIP(newNode, corev1.NodeExternalIP)

	clusterNode := &nodes.Node{
		NodeID:          machine.ID,
		PublicIPAddress: nodeIP,
		SSHUser:         sshUser,
		SSHKey:          sshkey,
	}

	return clusterNode, nil
}

// GetSSHUser gets the ssh user from a given clusterObject.
func GetSSHUser(client *rancher.Client, clusterObject *steveV1.SteveAPIObject) (string, error) {
	clusterSpec := &provv1.ClusterSpec{}
	err := steveV1.ConvertToK8sType(clusterObject.Spec, clusterSpec)
	if err != nil {
		return "", err
	}

	dynamicSchema := clusterSpec.RKEConfig.MachinePools[0].DynamicSchemaSpec
	var data clusters.DynamicSchemaSpec
	err = json.Unmarshal([]byte(dynamicSchema), &data)
	if err != nil {
		return "", err
	}

	sshUser := data.ResourceFields.SSHUser.Default.StringValue
	if sshUser == "" {
		sshUser = rootUser
	}

	return sshUser, nil
}
