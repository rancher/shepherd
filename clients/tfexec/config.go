package tfexec

import (
	"github.com/rancher/shepherd/pkg/config"
)

const (
	ConfigurationFileKey = "tfexec"
)

// A representation of the upstream PlanOptions that are typically used by this client
type PlanOptions struct {
	OutDir string `json:"outDir" yaml:"outDir"`
}

// A representative struct for the configuration options used by the tfexec client
type Config struct {
	WorkspaceName string       `json:"workspaceName" yaml:"workspaceName"`
	WorkingDir    string       `json:"workingDir" yaml:"workingDir"`
	ExecPath      string       `json:"execPath" yaml:"execPath"`
	VarFilePath   string       `json:"varFilePath" yaml:"varFilePath"`
	PlanFilePath  string       `json:"planFilePath" yaml:"planFilePath"`
	PlanOpts      *PlanOptions `json:"planOpts" yaml:"planOpts"`
}

// TerraformConfig loads the tfexec configuration inputs into a tfexec.Config struct and returns a pointer to it
func TerraformConfig() *Config {
	var tfConfig Config
	config.LoadConfig(ConfigurationFileKey, &tfConfig)
	return &tfConfig
}
