package cognito

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/rancher/shepherd/extensions/defaults"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

const (
	emailAttribute       = "email"
	cognitoIdentityStore = "COGNITO"
	classicHostedUI      = 1
	authorizePath        = "/oauth2/authorize"
)

// PasswordPolicy names the complexity a user pool requires of the passwords it accepts
type PasswordPolicy struct {
	MinimumLength    int32
	RequireLowercase bool
	RequireNumbers   bool
	RequireSymbols   bool
	RequireUppercase bool
}

// AppClient names the OAuth client a user pool serves the authorization code flow to
type AppClient struct {
	ID     string
	Secret string
}

// CreateUserPool creates a user pool and returns its identifier
func (c *Client) CreateUserPool(name string, passwordPolicy *PasswordPolicy) (string, error) {
	if name == "" {
		return "", fmt.Errorf("a Cognito user pool needs a name")
	}

	input := &cognitoidentityprovider.CreateUserPoolInput{
		PoolName:               aws.String(name),
		AutoVerifiedAttributes: []types.VerifiedAttributeType{types.VerifiedAttributeTypeEmail},
		AdminCreateUserConfig: &types.AdminCreateUserConfigType{
			AllowAdminCreateUserOnly: true,
		},
	}

	if passwordPolicy != nil {
		input.Policies = &types.UserPoolPolicyType{
			PasswordPolicy: &types.PasswordPolicyType{
				MinimumLength:    aws.Int32(passwordPolicy.MinimumLength),
				RequireLowercase: passwordPolicy.RequireLowercase,
				RequireNumbers:   passwordPolicy.RequireNumbers,
				RequireSymbols:   passwordPolicy.RequireSymbols,
				RequireUppercase: passwordPolicy.RequireUppercase,
			},
		}
	}

	ctx, cancel := callContext()
	defer cancel()

	created, err := c.svc.CreateUserPool(ctx, input)
	if err != nil {
		return "", fmt.Errorf("creating the %s user pool: %w", name, err)
	}

	userPoolID := aws.ToString(created.UserPool.Id)

	if c.Session != nil {
		c.Session.RegisterCleanupFunc(func() error {
			return c.DeleteUserPool(userPoolID)
		})
	}

	return userPoolID, nil
}

// DeleteUserPool removes a user pool
func (c *Client) DeleteUserPool(userPoolID string) error {
	ctx, cancel := callContext()
	defer cancel()

	if _, err := c.svc.DeleteUserPool(ctx, &cognitoidentityprovider.DeleteUserPoolInput{
		UserPoolId: aws.String(userPoolID),
	}); err != nil && !isNotFound(err) {
		return fmt.Errorf("deleting the %s user pool: %w", userPoolID, err)
	}

	return nil
}

// CreateUserPoolDomain creates the hosted UI domain a user pool serves its sign-in form from, and waits for it to serve
func (c *Client) CreateUserPoolDomain(userPoolID, domainPrefix string) error {
	if err := requireUserPool(userPoolID, "create the "+domainPrefix+" hosted UI domain"); err != nil {
		return err
	}

	ctx, cancel := callContext()
	defer cancel()

	if _, err := c.svc.CreateUserPoolDomain(ctx,
		&cognitoidentityprovider.CreateUserPoolDomainInput{
			Domain:              aws.String(domainPrefix),
			UserPoolId:          aws.String(userPoolID),
			ManagedLoginVersion: aws.Int32(classicHostedUI),
		}); err != nil {
		return fmt.Errorf("creating the %s hosted UI domain: %w", domainPrefix, err)
	}

	if c.Session != nil {
		c.Session.RegisterCleanupFunc(func() error {
			return c.DeleteUserPoolDomain(userPoolID, domainPrefix)
		})
	}

	if err := c.waitForDomain(userPoolID, domainPrefix, types.DomainStatusTypeActive); err != nil {
		return err
	}

	return c.waitForDomainServing(c.HostedUIDomain(domainPrefix))
}

func (c *Client) waitForDomainServing(domain string) error {
	address, err := url.JoinPath(domain, authorizePath)
	if err != nil {
		return fmt.Errorf("building the authorize address of %s: %w", domain, err)
	}

	httpClient := &http.Client{Timeout: defaults.OneMinuteTimeout}

	var lastErr error

	if err := kwait.PollUntilContextTimeout(context.Background(), defaults.TenSecondTimeout,
		defaults.TenMinuteTimeout, false, func(ctx context.Context) (bool, error) {
			request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
			if err != nil {
				return false, err
			}

			response, err := httpClient.Do(request)
			if err != nil {
				lastErr = err

				return false, nil
			}

			defer response.Body.Close()

			if response.StatusCode >= http.StatusInternalServerError {
				lastErr = fmt.Errorf("status %d", response.StatusCode)

				return false, nil
			}

			return true, nil
		}); err != nil {
		return fmt.Errorf("waiting for %s to answer, Cognito reports a hosted UI domain active before the "+
			"distribution behind it resolves: %w (last attempt: %v)", address, err, lastErr)
	}

	return nil
}

// DeleteUserPoolDomain removes a hosted UI domain from a user pool and waits for it to go
func (c *Client) DeleteUserPoolDomain(userPoolID, domainPrefix string) error {
	ctx, cancel := callContext()
	defer cancel()

	if _, err := c.svc.DeleteUserPoolDomain(ctx,
		&cognitoidentityprovider.DeleteUserPoolDomainInput{
			Domain:     aws.String(domainPrefix),
			UserPoolId: aws.String(userPoolID),
		}); err != nil && !isNotFound(err) {
		return fmt.Errorf("deleting the %s hosted UI domain: %w", domainPrefix, err)
	}

	return c.waitForDomain(userPoolID, domainPrefix, "")
}

func (c *Client) waitForDomain(userPoolID, domainPrefix string, want types.DomainStatusType) error {
	var lastStatus types.DomainStatusType

	err := kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout,
		defaults.FiveMinuteTimeout, true, func(ctx context.Context) (bool, error) {
			described, err := c.svc.DescribeUserPoolDomain(ctx,
				&cognitoidentityprovider.DescribeUserPoolDomainInput{Domain: aws.String(domainPrefix)})
			if err != nil {
				if isNotFound(err) {
					return want == "", nil
				}

				return false, err
			}

			if described.DomainDescription == nil || described.DomainDescription.Domain == nil {
				return want == "", nil
			}

			lastStatus = described.DomainDescription.Status
			if lastStatus == types.DomainStatusTypeFailed {
				return false, fmt.Errorf("the %s hosted UI domain of user pool %s reached %s",
					domainPrefix, userPoolID, lastStatus)
			}

			return lastStatus == want, nil
		})
	if kwait.Interrupted(err) {
		return fmt.Errorf("the %s hosted UI domain of user pool %s never reached %q, its last status was %q: %w",
			domainPrefix, userPoolID, want, lastStatus, err)
	}

	if err != nil {
		return fmt.Errorf("waiting for the %s hosted UI domain of user pool %s: %w", domainPrefix, userPoolID, err)
	}

	return nil
}

// CreateUserPoolClient creates an app client that completes the authorization code flow against a user pool
func (c *Client) CreateUserPoolClient(userPoolID, clientName string, callbackURLs, logoutURLs,
	scopes []string) (*AppClient, error) {
	if len(callbackURLs) == 0 {
		return nil, fmt.Errorf("the %s app client needs at least one callback URL the authorization code is "+
			"sent back to", clientName)
	}

	if err := requireUserPool(userPoolID, "create the "+clientName+" app client"); err != nil {
		return nil, err
	}

	ctx, cancel := callContext()
	defer cancel()

	created, err := c.svc.CreateUserPoolClient(ctx,
		&cognitoidentityprovider.CreateUserPoolClientInput{
			UserPoolId:                      aws.String(userPoolID),
			ClientName:                      aws.String(clientName),
			GenerateSecret:                  true,
			CallbackURLs:                    callbackURLs,
			LogoutURLs:                      logoutURLs,
			AllowedOAuthFlows:               []types.OAuthFlowType{types.OAuthFlowTypeCode},
			AllowedOAuthScopes:              scopes,
			AllowedOAuthFlowsUserPoolClient: true,
			SupportedIdentityProviders:      []string{cognitoIdentityStore},
		})
	if err != nil {
		return nil, fmt.Errorf("creating the %s app client in user pool %s: %w", clientName, userPoolID, err)
	}

	appClient := &AppClient{
		ID:     aws.ToString(created.UserPoolClient.ClientId),
		Secret: aws.ToString(created.UserPoolClient.ClientSecret),
	}

	if c.Session != nil {
		c.Session.RegisterCleanupFunc(func() error {
			return c.DeleteUserPoolClient(userPoolID, appClient.ID)
		})
	}

	return appClient, nil
}

// DeleteUserPoolClient removes an app client from a user pool
func (c *Client) DeleteUserPoolClient(userPoolID, clientID string) error {
	ctx, cancel := callContext()
	defer cancel()

	if _, err := c.svc.DeleteUserPoolClient(ctx,
		&cognitoidentityprovider.DeleteUserPoolClientInput{
			UserPoolId: aws.String(userPoolID),
			ClientId:   aws.String(clientID),
		}); err != nil && !isNotFound(err) {
		return fmt.Errorf("deleting the %s app client from user pool %s: %w", clientID, userPoolID, err)
	}

	return nil
}

// IssuerOf returns the address Cognito issues the pool's tokens under
func (c *Client) IssuerOf(userPoolID string) string {
	return fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", c.Config.Region, userPoolID)
}

// HostedUIDomain returns the address the hosted sign-in form and the OAuth endpoints are served from
func (c *Client) HostedUIDomain(domainPrefix string) string {
	return fmt.Sprintf("https://%s.auth.%s.amazoncognito.com", domainPrefix, c.Config.Region)
}
