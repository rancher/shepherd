package k3d

// The json/yaml config key for the k3d config
const ConfigurationFileKey = "k3d"

// Config is configuration needed to create k3d cluster for integration testing.
type Config struct {
	createTimeout int `yaml:"createTimeout" default:"120s"`
}
