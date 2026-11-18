package cognito

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

const (
	subAttribute           = "sub"
	emailVerifiedAttribute = "email_verified"
)

// User represents a Cognito account together with the credentials it signs in with
type User struct {
	Username string
	Password string
	Email    string
	Sub      string
}

// CreateUser creates a confirmed account with a permanent password in the user pool
func (c *Client) CreateUser(userPoolID, username, email, password string) (*User, error) {
	if err := requireUserPool(userPoolID, "create "+username); err != nil {
		return nil, err
	}

	ctx, cancel := callContext()
	defer cancel()

	created, err := c.svc.AdminCreateUser(ctx, &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId:    aws.String(userPoolID),
		Username:      aws.String(username),
		MessageAction: types.MessageActionTypeSuppress,
		UserAttributes: []types.AttributeType{
			{Name: aws.String(emailAttribute), Value: aws.String(email)},
			{Name: aws.String(emailVerifiedAttribute), Value: aws.String("true")},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("creating the %s account in user pool %s: %w", username, userPoolID, err)
	}

	subject := attributeValue(created.User.Attributes, subAttribute)
	if subject == "" {
		return nil, fmt.Errorf("Cognito created %s but asserted no %s attribute, which is the immutable "+
			"identifier Rancher builds the principal from", username, subAttribute)
	}

	if c.Session != nil {
		c.Session.RegisterCleanupFunc(func() error {
			return c.deleteUser(userPoolID, username)
		})
	}

	passwordCtx, passwordCancel := callContext()
	defer passwordCancel()

	if _, err := c.svc.AdminSetUserPassword(passwordCtx, &cognitoidentityprovider.AdminSetUserPasswordInput{
		UserPoolId: aws.String(userPoolID),
		Username:   aws.String(username),
		Password:   aws.String(password),
		Permanent:  true,
	}); err != nil {
		return nil, fmt.Errorf("setting a permanent password on %s, without one Cognito answers the sign-in "+
			"with a password change challenge: %w", username, err)
	}

	return &User{Username: username, Password: password, Email: email, Sub: subject}, nil
}

func (c *Client) deleteUser(userPoolID, username string) error {
	ctx, cancel := callContext()
	defer cancel()

	if _, err := c.svc.AdminDeleteUser(ctx, &cognitoidentityprovider.AdminDeleteUserInput{
		UserPoolId: aws.String(userPoolID),
		Username:   aws.String(username),
	}); err != nil && !isNotFound(err) {
		return fmt.Errorf("deleting the %s account from user pool %s: %w", username, userPoolID, err)
	}

	return nil
}

func attributeValue(attributes []types.AttributeType, name string) string {
	for _, attribute := range attributes {
		if aws.ToString(attribute.Name) == name {
			return aws.ToString(attribute.Value)
		}
	}

	return ""
}
