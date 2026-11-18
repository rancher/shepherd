package cognito

const ConfigurationFileKey = "cognito"

const (
	defaultRegion          = "us-east-2"
	defaultUserEmailDomain = "rancher.test"
)

// Config represents the AWS Cognito account and user pool a run administers
type Config struct {
	Region          string `json:"region" yaml:"region"`
	AccessKey       string `json:"accessKey" yaml:"accessKey"`
	SecretKey       string `json:"secretKey" yaml:"secretKey"`
	SessionToken    string `json:"sessionToken" yaml:"sessionToken"`
	UserPoolName    string `json:"userPoolName" yaml:"userPoolName"`
	DomainPrefix    string `json:"domainPrefix" yaml:"domainPrefix"`
	UserEmailDomain string `json:"userEmailDomain" yaml:"userEmailDomain"`
}

func (c *Config) applyDefaults() {
	if c.Region == "" {
		c.Region = defaultRegion
	}

	if c.UserEmailDomain == "" {
		c.UserEmailDomain = defaultUserEmailDomain
	}
}
