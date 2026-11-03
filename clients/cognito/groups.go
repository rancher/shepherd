package cognito

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

// CreateGroup creates a group in the user pool
func (c *Client) CreateGroup(name string) error {
	if c.userPoolID == "" {
		return fmt.Errorf("no user pool is known to this client, create one before adding the %s group to it", name)
	}

	if _, err := c.svc.CreateGroup(context.Background(), &cognitoidentityprovider.CreateGroupInput{
		UserPoolId: aws.String(c.userPoolID),
		GroupName:  aws.String(name),
	}); err != nil {
		return fmt.Errorf("creating the %s group in user pool %s: %w", name, c.userPoolID, err)
	}

	if c.Session != nil {
		userPoolID := c.userPoolID

		c.Session.RegisterCleanupFunc(func() error {
			return c.deleteGroup(userPoolID, name)
		})
	}

	return nil
}

// DeleteGroup removes a group from the user pool
func (c *Client) DeleteGroup(name string) error {
	return c.deleteGroup(c.userPoolID, name)
}

func (c *Client) deleteGroup(userPoolID, name string) error {
	if _, err := c.svc.DeleteGroup(context.Background(), &cognitoidentityprovider.DeleteGroupInput{
		UserPoolId: aws.String(userPoolID),
		GroupName:  aws.String(name),
	}); err != nil && !isNotFound(err) {
		return fmt.Errorf("deleting the %s group from user pool %s: %w", name, userPoolID, err)
	}

	return nil
}

// AddUserToGroup makes an account a member of a group
func (c *Client) AddUserToGroup(username, group string) error {
	if _, err := c.svc.AdminAddUserToGroup(context.Background(),
		&cognitoidentityprovider.AdminAddUserToGroupInput{
			UserPoolId: aws.String(c.userPoolID),
			Username:   aws.String(username),
			GroupName:  aws.String(group),
		}); err != nil {
		return fmt.Errorf("adding %s to the %s group: %w", username, group, err)
	}

	return nil
}
