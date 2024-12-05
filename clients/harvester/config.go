package harvester

// The json/yaml config key for the harvester config
const ConfigurationFileKey = "harvester"

// Config is configuration need to test against a harvester instance
type Config struct {
	Host          string `yaml:"host" json:"host"`
	AdminToken    string `yaml:"adminToken" json:"adminToken"`
	AdminPassword string `yaml:"adminPassword" json:"adminPassword"`
	Insecure      *bool  `yaml:"insecure" json:"insecure" default:"true"`
	Cleanup       *bool  `yaml:"cleanup" json:"cleanup" default:"true"`
}
