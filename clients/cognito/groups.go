package cognito

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

// CreateGroup creates a group in the user pool
func (c *Client) CreateGroup(userPoolID, name string) error {
	if err := requireUserPool(userPoolID, "create the "+name+" group"); err != nil {
		return err
	}

	ctx, cancel := callContext()
	defer cancel()

	if _, err := c.svc.CreateGroup(ctx, &cognitoidentityprovider.CreateGroupInput{
		UserPoolId: aws.String(userPoolID),
		GroupName:  aws.String(name),
	}); err != nil {
		return fmt.Errorf("creating the %s group in user pool %s: %w", name, userPoolID, err)
	}

	if c.Session != nil {
		c.Session.RegisterCleanupFunc(func() error {
			return c.deleteGroup(userPoolID, name)
		})
	}

	return nil
}

// AddUserToGroup makes an account a member of a group in the user pool
func (c *Client) AddUserToGroup(userPoolID, username, group string) error {
	if err := requireUserPool(userPoolID, "add "+username+" to the "+group+" group"); err != nil {
		return err
	}

	ctx, cancel := callContext()
	defer cancel()

	if _, err := c.svc.AdminAddUserToGroup(ctx, &cognitoidentityprovider.AdminAddUserToGroupInput{
		UserPoolId: aws.String(userPoolID),
		Username:   aws.String(username),
		GroupName:  aws.String(group),
	}); err != nil {
		return fmt.Errorf("adding %s to the %s group in user pool %s: %w", username, group, userPoolID, err)
	}

	return nil
}

func (c *Client) deleteGroup(userPoolID, name string) error {
	ctx, cancel := callContext()
	defer cancel()

	if _, err := c.svc.DeleteGroup(ctx, &cognitoidentityprovider.DeleteGroupInput{
		UserPoolId: aws.String(userPoolID),
		GroupName:  aws.String(name),
	}); err != nil && !isNotFound(err) {
		return fmt.Errorf("deleting the %s group from user pool %s: %w", name, userPoolID, err)
	}

	return nil
}
