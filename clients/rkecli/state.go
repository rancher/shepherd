package rkecli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/rke/cluster"
	rketypes "github.com/rancher/rke/types"
	"github.com/rancher/shepherd/clients/rancher"
	v3 "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	v1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/file"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
)

const filePrefix = "cluster"
const dirName = "rke-cattle-test-dir"

// NewRKEConfigs creates a new dir.
// In that dir, it generates state and cluster files from the state configmap.
// Returns generated state and cluster files' paths.
func NewRKEConfigs(client *rancher.Client) (stateFilePath, clusterFilePath string, err error) {
	rkeConfig := new(Config)
	config.LoadConfig(ConfigurationFileKey, rkeConfig)

	err = file.NewDir(dirName)
	if err != nil {
		return
	}

	state, err := GetFullState(client)
	if err != nil {
		return
	}

	clusterFilePath, err = NewClusterFile(state, dirName, rkeConfig)
	if err != nil {
		return
	}

	stateFilePath, err = NewStateFile(state, dirName)
	if err != nil {
		return
	}

	return
}

// ReadClusterFromStateFile is a function that reads the RKE config from the given state file path.
// Returns RKE config.
func ReadClusterFromStateFile(stateFilePath string) (*v3.RancherKubernetesEngineConfig, error) {
	byteState, err := os.ReadFile(stateFilePath)
	if err != nil {
		return nil, err
	}

	//get bytes state
	state := make(map[string]interface{})
	err = json.Unmarshal(byteState, &state)
	if err != nil {
		return nil, err
	}

	//reach rke config in the current state
	byteRkeConfig := state["currentState"].(map[string]interface{})["rkeConfig"]
	byteTest, err := json.Marshal(byteRkeConfig)
	if err != nil {
		return nil, err
	}

	//final unmarshal to get the struct
	rkeConfig := new(v3.RancherKubernetesEngineConfig)
	err = json.Unmarshal(byteTest, rkeConfig)
	if err != nil {
		return nil, err
	}

	return rkeConfig, nil
}

// UpdateKubernetesVersion is a function that updates kubernetes version value in cluster.yml file.
func UpdateKubernetesVersion(kubernetesVersion, clusterFilePath string) error {
	byteRkeConfig, err := os.ReadFile(clusterFilePath)
	if err != nil {
		return err
	}

	rkeConfig := new(rketypes.RancherKubernetesEngineConfig)
	err = yaml.Unmarshal(byteRkeConfig, rkeConfig)
	if err != nil {
		return err
	}

	rkeConfig.Version = kubernetesVersion

	byteConfig, err := yaml.Marshal(rkeConfig)
	if err != nil {
		return err
	}

	return os.WriteFile(clusterFilePath, byteConfig, 0644)
}

// NewClusterFile is a function that generates new cluster.yml file from the full state.
// Returns the generated file's path.
func NewClusterFile(state *cluster.FullState, dirName string, config *Config) (clusterFilePath string, err error) {
	extension := "yml"
	rkeConfigFileName := fmt.Sprintf("%v/%v.%v", dirName, filePrefix, extension)

	rkeConfig := rketypes.RancherKubernetesEngineConfig{}
	currentRkeConfig := state.CurrentState.RancherKubernetesEngineConfig.DeepCopy()

	rkeConfig.Version = currentRkeConfig.Version
	rkeConfig.Nodes = currentRkeConfig.Nodes

	if config.SSHKey != "" {
		for i := range rkeConfig.Nodes {
			rkeConfig.Nodes[i].SSHKey = config.SSHKey
		}
	} else if config.SSHPath != "" {
		rkeConfig.SSHKeyPath = appendSSHPath(currentRkeConfig.SSHKeyPath, config.SSHPath)
		for i := range rkeConfig.Nodes {
			rkeConfig.Nodes[i].SSHKeyPath = appendSSHPath(rkeConfig.Nodes[i].SSHKeyPath, config.SSHPath)
		}
	} else {
		return "", errors.New("missing SSH configuration")
	}

	marshaled, err := yaml.Marshal(rkeConfig)
	if err != nil {
		return
	}

	fileName := file.Name(rkeConfigFileName)

	clusterFilePath, err = fileName.NewFile(marshaled)
	if err != nil {
		return
	}

	return
}

// NewStateFile is a function that generates new cluster.rkestate file from the full state.
// Returns the generated file's path.
func NewStateFile(state *cluster.FullState, dirName string) (stateFilePath string, err error) {
	extension := "rkestate"
	rkeStateFileName := fmt.Sprintf("%v/%v.%v", dirName, filePrefix, extension)

	marshaled, err := json.Marshal(state)
	if err != nil {
		return
	}

	stateFilePath, err = file.Name(rkeStateFileName).NewFile(marshaled)
	if err != nil {
		return
	}

	return
}

// GetFullState is a function that gets RKE full state from "full-cluster-state" secret.
// And returns the cluster full state.
func GetFullState(client *rancher.Client) (state *cluster.FullState, err error) {
	namespacedSecretClient := client.Steve.SteveType("secret").NamespacedSteveClient(cluster.SystemNamespace)

	fullstateSecretID := fmt.Sprintf("%s/%s", cluster.SystemNamespace, cluster.FullStateSecretName)

	secretResp, err := namespacedSecretClient.ByID(fullstateSecretID)
	if err != nil {
		return
	}

	secret := &corev1.Secret{}
	err = v1.ConvertToK8sType(secretResp.JSONResp, secret)
	if err != nil {
		return
	}

	rawState, ok := secret.Data[cluster.FullStateSecretName]
	if !ok {
		err = errors.Wrapf(err, "couldn't retrieve full state data in the secret")
		return
	}

	rkeFullState := &cluster.FullState{}
	err = json.Unmarshal([]byte(rawState), rkeFullState)
	if err != nil {
		return
	}

	return rkeFullState, nil
}

// appendSSHPath reads sshPath input from the cattle config file.
// If the config input has a different prefix, adds the prefix.
func appendSSHPath(currentPath, newPath string) string {
	if strings.HasPrefix(currentPath, newPath) {
		return currentPath
	}

	ssh := ".ssh/"
	currentPath = strings.TrimPrefix(currentPath, ssh)

	return fmt.Sprintf("%s/%s", newPath, currentPath)
}
