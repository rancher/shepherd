package ec2

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/rancher/shepherd/extensions/cloudcredentials"
	"github.com/rancher/shepherd/pkg/config"
)

// Client is a struct that wraps the needed AWSEC2Config object, and ec2.EC2 which makes the actual calls to aws.
type Client struct {
	SVC          *ec2.EC2
	ClientConfig *AWSEC2Configs
}

// NewClient is a constructor that creates an *Client which a wrapper for a "github.com/aws/aws-sdk-go/service/ec2" session and
// the aws ec2 config.
func NewClient() (*Client, error) {
	awsEC2ClientConfig := new(AWSEC2Configs)

	config.LoadConfig(ConfigurationFileKey, awsEC2ClientConfig)

	credential := credentials.NewStaticCredentials(awsEC2ClientConfig.AWSAccessKeyID, awsEC2ClientConfig.AWSSecretAccessKey, "")
	sess, err := session.NewSession(&aws.Config{
		Credentials: credential,
		Region:      aws.String(awsEC2ClientConfig.Region)},
	)
	if err != nil {
		return nil, err
	}

	svc := ec2.New(sess)
	return &Client{
		SVC:          svc,
		ClientConfig: awsEC2ClientConfig,
	}, nil
}

// NewClientFromConfig generates a new ec2 client using the provided credentials
func NewClientFromConfig(awsCredentials *cloudcredentials.AmazonEC2CredentialConfig) (*Client, error) {
	credential := credentials.NewStaticCredentials(awsCredentials.AccessKey, awsCredentials.SecretKey, "")
	ec2Session, err := session.NewSession(&aws.Config{
		Credentials: credential,
		Region:      aws.String(awsCredentials.DefaultRegion)},
	)
	if err != nil {
		return nil, err
	}

	svc := ec2.New(ec2Session)
	return &Client{
		SVC: svc,
		ClientConfig: &AWSEC2Configs{
			AWSAccessKeyID:     awsCredentials.AccessKey,
			AWSSecretAccessKey: awsCredentials.SecretKey,
			Region:             awsCredentials.DefaultRegion,
		},
	}, nil
}
