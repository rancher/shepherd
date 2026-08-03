package oidc

const (
	ConfigurationFileKey                 = "oidc"
	OIDCProviderFeatureFlag              = "oidc-provider"
	DefaultTokenExpirationSeconds        = 3600
	DefaultRefreshTokenExpirationSeconds = 86400
)

var DefaultAutomationScopes = []string{
	"openid",
	"profile",
	"offline_access",
	"rancher:users",
}

type Config struct {
	ClientName                    string   `json:"clientName" yaml:"clientName"`
	RedirectURI                   string   `json:"redirectURI" yaml:"redirectURI"`
	Scopes                        []string `json:"scopes" yaml:"scopes"`
	TokenExpirationSeconds        int      `json:"tokenExpirationSeconds" yaml:"tokenExpirationSeconds"`
	RefreshTokenExpirationSeconds int      `json:"refreshTokenExpirationSeconds" yaml:"refreshTokenExpirationSeconds"`
	AdminUsername                 string   `json:"adminUsername" yaml:"adminUsername"`
	AdminPassword                 string   `json:"adminPassword" yaml:"adminPassword"`
}
