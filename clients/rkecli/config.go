package rkecli

const ConfigurationFileKey = "rke"

// RKE configuration required to run rkecli and up
type Config struct {
	SSHKey  string `json:"sshKey,omitempty" yaml:"sshKey,omitempty" default:""`
	SSHPath string `json:"sshPath,omitempty" yaml:"sshPath,omitempty" default:""`
}
