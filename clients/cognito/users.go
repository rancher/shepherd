package cognito

import (
	"context"
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

// CreateUser creates a confirmed account with a permanent password
func (c *Client) CreateUser(username, email, password string) (*User, error) {
	if c.userPoolID == "" {
		return nil, fmt.Errorf("no user pool is known to this client, create one before adding %s to it", username)
	}

	created, err := c.svc.AdminCreateUser(context.Background(), &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId:    aws.String(c.userPoolID),
		Username:      aws.String(username),
		MessageAction: types.MessageActionTypeSuppress,
		UserAttributes: []types.AttributeType{
			{Name: aws.String(emailAttribute), Value: aws.String(email)},
			{Name: aws.String(emailVerifiedAttribute), Value: aws.String("true")},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("creating the %s account in user pool %s: %w", username, c.userPoolID, err)
	}

	subject := attributeValue(created.User.Attributes, subAttribute)
	if subject == "" {
		return nil, fmt.Errorf("Cognito created %s but asserted no %s attribute, which is the immutable "+
			"identifier Rancher builds the principal from", username, subAttribute)
	}

	if c.Session != nil {
		userPoolID := c.userPoolID

		c.Session.RegisterCleanupFunc(func() error {
			return c.deleteUser(userPoolID, username)
		})
	}

	if _, err := c.svc.AdminSetUserPassword(context.Background(),
		&cognitoidentityprovider.AdminSetUserPasswordInput{
			UserPoolId: aws.String(c.userPoolID),
			Username:   aws.String(username),
			Password:   aws.String(password),
			Permanent:  true,
		}); err != nil {
		return nil, fmt.Errorf("setting a permanent password on %s, without one Cognito answers the sign-in "+
			"with a password change challenge: %w", username, err)
	}

	return &User{Username: username, Password: password, Email: email, Sub: subject}, nil
}

// DeleteUser removes an account from the user pool
func (c *Client) DeleteUser(username string) error {
	return c.deleteUser(c.userPoolID, username)
}

func (c *Client) deleteUser(userPoolID, username string) error {
	if _, err := c.svc.AdminDeleteUser(context.Background(), &cognitoidentityprovider.AdminDeleteUserInput{
		UserPoolId: aws.String(userPoolID),
		Username:   aws.String(username),
	}); err != nil && !isNotFound(err) {
		return fmt.Errorf("deleting the %s account from user pool %s: %w", username, userPoolID, err)
	}

	return nil
}

// GetUser returns the account the given username names, and nil when the pool holds no such account
func (c *Client) GetUser(username string) (*User, error) {
	found, err := c.svc.AdminGetUser(context.Background(), &cognitoidentityprovider.AdminGetUserInput{
		UserPoolId: aws.String(c.userPoolID),
		Username:   aws.String(username),
	})
	if isNotFound(err) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("retrieving the %s account from user pool %s: %w", username, c.userPoolID, err)
	}

	return &User{
		Username: aws.ToString(found.Username),
		Email:    attributeValue(found.UserAttributes, emailAttribute),
		Sub:      attributeValue(found.UserAttributes, subAttribute),
	}, nil
}

func attributeValue(attributes []types.AttributeType, name string) string {
	for _, attribute := range attributes {
		if aws.ToString(attribute.Name) == name {
			return aws.ToString(attribute.Value)
		}
	}

	return ""
}
