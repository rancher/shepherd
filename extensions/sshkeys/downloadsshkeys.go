package sshkeys

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/rancher/shepherd/clients/rancher"
	steveV1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	"github.com/rancher/shepherd/extensions/defaults/stevetypes"
	"github.com/rancher/shepherd/pkg/nodes"
)

const (
	ClusterMachineAnnotation = "cluster.x-k8s.io/machine"
	sshConfigFile            = "config.json"
	sshKeyFile               = "id_rsa"
)

// DownloadSSHCredentials is a helper function that downloads SSH credentials from a rancher machine and returns the SSHKey, Username and IP
func DownloadSSHCredentials(client *rancher.Client, machinePoolNodeName string) (key, username, ip string, err error) {
	machine, err := client.Steve.SteveType(stevetypes.Machine).ByID(namespaces.FleetDefault + "/" + machinePoolNodeName)
	if err != nil {
		return "", "", "", err
	}

	sshKeyLink := machine.Links["sshkeys"]

	req, err := http.NewRequest("GET", sshKeyLink, nil)
	if err != nil {
		return "", "", "", err
	}

	req.Header.Add("Authorization", "Bearer "+client.RancherConfig.AdminToken)

	resp, err := client.Management.APIBaseClient.Ops.Client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(bodyBytes), int64(len(bodyBytes)))
	if err != nil {
		return "", "", "", err
	}

	var sshConfigJson []byte
	var sshKey []byte
	for _, file := range zipReader.File {
		rc, err := file.Open()
		if err != nil {
			return "", "", "", err
		}
		defer rc.Close()

		content, err := io.ReadAll(rc)
		if err != nil {
			return "", "", "", err
		}

		if machinePoolNodeName+"/"+sshConfigFile == file.Name {
			sshConfigJson = content
		} else if machinePoolNodeName+"/"+sshKeyFile == file.Name {
			sshKey = content
		}
	}

	var sshConfig map[string]any
	err = json.Unmarshal(sshConfigJson, &sshConfig)
	if err != nil {
		return "", "", "", err
	}

	return string(sshKey), sshConfig["SSHUser"].(string), sshConfig["IPAddress"].(string), err
}

// GetSSHNodeFromMachine returns the v1/node object given a steve/v1/machine object.
func GetSSHNodeFromMachine(client *rancher.Client, machine *steveV1.SteveAPIObject) (*nodes.Node, error) {
	machineName := machine.Annotations[ClusterMachineAnnotation]
	sshKey, sshUser, sshIPAddress, err := DownloadSSHCredentials(client, machineName)
	if err != nil {
		return nil, err
	}

	clusterNode := &nodes.Node{
		NodeID:          machine.ID,
		PublicIPAddress: sshIPAddress,
		SSHUser:         sshUser,
		SSHKey:          []byte(sshKey),
	}

	return clusterNode, nil
}
