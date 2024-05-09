package provisioninginput

import (
	"path"
	"runtime"

	"github.com/rancher/shepherd/extensions/defaults"
)

const (
	machinePoolsFile = "machinepools.yml"
	nodePoolsFile    = "nodepools.yml"
)

func GetMachinePoolConfigs(machinePoolConfigNames []string) []MachinePools {
	machinePoolConfigs := []MachinePools{}
	_, filename, _, _ := runtime.Caller(0)
	file := path.Join(path.Dir(filename), machinePoolsFile)

	for _, name := range machinePoolConfigNames {
		myconfig := new(MachinePools)
		defaults.LoadDefault(file, name, myconfig)
		machinePoolConfigs = append(machinePoolConfigs, *myconfig)
	}

	return machinePoolConfigs
}

func GetNodePoolConfigs(nodePoolConfigNames []string) []NodePools {
	nodePoolConfigs := []NodePools{}
	_, filename, _, _ := runtime.Caller(0)
	file := path.Join(path.Dir(filename), nodePoolsFile)

	for _, name := range nodePoolConfigNames {
		myconfig := new(NodePools)
		defaults.LoadDefault(file, name, myconfig)
		nodePoolConfigs = append(nodePoolConfigs, *myconfig)
	}

	return nodePoolConfigs
}
