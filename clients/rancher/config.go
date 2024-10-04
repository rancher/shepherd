package rancher

// The json/yaml config key for the rancher config
const ConfigurationFileKey = "rancher"

// ConfigCommon contains the common fields for both Config and InstanceConfig
type ConfigCommon struct {
	Host          string `yaml:"host" json:"host"`
	AdminToken    string `yaml:"adminToken" json:"adminToken"`
	AdminPassword string `yaml:"adminPassword" json:"adminPassword"`
	Insecure      *bool  `yaml:"insecure" json:"insecure" default:"true"`
	Cleanup       *bool  `yaml:"cleanup" json:"cleanup" default:"true"`
	CAFile        string `yaml:"caFile" json:"caFile" default:""`
	CACerts       string `yaml:"caCerts" json:"caCerts" default:""`
	ClusterName   string `yaml:"clusterName" json:"clusterName" default:""`
	ShellImage    string `yaml:"shellImage" json:"shellImage" default:""`
	RancherCLI    bool   `yaml:"rancherCLI" json:"rancherCLI" default:"false"`
}

// Config is configuration needed to test it against a rancher instance
type Config struct {
	ConfigCommon     `yaml:",inline"`
	RancherInstances []InstanceConfig `yaml:"rancherInstances" json:"rancherInstances" default:"[]"`
}

// InstanceConfig represents configuration for additional Rancher instances
type InstanceConfig struct {
	ConfigCommon `yaml:",inline"`
}
