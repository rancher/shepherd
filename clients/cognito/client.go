package cognito

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
)

type Client struct {
	Config  *Config
	Session *session.Session

	svc        *cognitoidentityprovider.Client
	userPoolID string
	domain     string
}

// NewClient constructs a Cognito administration client from an already loaded configuration
func NewClient(cognitoConfig *Config, testSession *session.Session) (*Client, error) {
	if cognitoConfig == nil {
		return nil, fmt.Errorf("a Cognito client needs a config")
	}

	cognitoConfig.applyDefaults()

	if cognitoConfig.AccessKey == "" || cognitoConfig.SecretKey == "" {
		return nil, fmt.Errorf("accessKey and secretKey must both be set under the %s config key, the client "+
			"administers the user pool on the caller's behalf", ConfigurationFileKey)
	}

	awsConfig := aws.Config{
		Region: cognitoConfig.Region,
		Credentials: credentials.NewStaticCredentialsProvider(cognitoConfig.AccessKey, cognitoConfig.SecretKey,
			cognitoConfig.SessionToken),
	}

	return &Client{
		Config:     cognitoConfig,
		Session:    testSession,
		svc:        cognitoidentityprovider.NewFromConfig(awsConfig),
		userPoolID: cognitoConfig.UserPoolID,
	}, nil
}

// NewClientFromConfigKey constructs a Cognito administration client from the given configuration key
func NewClientFromConfigKey(configKey string, testSession *session.Session) (*Client, error) {
	cognitoConfig := new(Config)
	config.LoadConfig(configKey, cognitoConfig)

	return NewClient(cognitoConfig, testSession)
}

// UserPoolID returns the identifier of the user pool this client administers
func (c *Client) UserPoolID() string {
	return c.userPoolID
}

// Domain returns the hosted UI domain the authorization code flow is served from
func (c *Client) Domain() string {
	return c.domain
}

func isNotFound(err error) bool {
	var resourceNotFound *types.ResourceNotFoundException
	var userNotFound *types.UserNotFoundException

	return errors.As(err, &resourceNotFound) || errors.As(err, &userNotFound)
}
