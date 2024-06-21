package vai

const (
	ConfigurationFileKey = "vai"
)

type Config struct {
	Enabled bool `json:"enabled" yaml:"enabled" default:"false"`
}
