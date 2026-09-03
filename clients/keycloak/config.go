package keycloak

const ConfigurationFileKey = "keycloak"

const (
	defaultAdminRealm      = "master"
	defaultAdminClientID   = "admin-cli"
	defaultUserEmailDomain = "rancher.test"
)

// Config represents the Keycloak server configuration structure
type Config struct {
	Host            string `json:"keycloakHost" yaml:"keycloakHost"`
	Realm           string `json:"keycloakRealm" yaml:"keycloakRealm"`
	AdminRealm      string `json:"keycloakAdminRealm" yaml:"keycloakAdminRealm" default:"master"`
	AdminClientID   string `json:"keycloakAdminClientID" yaml:"keycloakAdminClientID" default:"admin-cli"`
	AdminUser       string `json:"keycloakAdminUser" yaml:"keycloakAdminUser"`
	AdminPassword   string `json:"keycloakAdminPassword" yaml:"keycloakAdminPassword"`
	Insecure        bool   `json:"keycloakInsecure" yaml:"keycloakInsecure"`
	UserEmailDomain string `json:"keycloakUserEmailDomain" yaml:"keycloakUserEmailDomain" default:"rancher.test"`
}

func (c *Config) applyDefaults() {
	if c.AdminRealm == "" {
		c.AdminRealm = defaultAdminRealm
	}

	if c.AdminClientID == "" {
		c.AdminClientID = defaultAdminClientID
	}

	if c.UserEmailDomain == "" {
		c.UserEmailDomain = defaultUserEmailDomain
	}
}
