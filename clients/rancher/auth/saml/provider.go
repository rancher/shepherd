package saml

import (
	"context"
	"fmt"
	"strings"

	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	samlext "github.com/rancher/shepherd/extensions/auth/saml"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

const serviceProviderWarmStreak = 12

type ProviderOperations interface {
	Enable() error
	Disable() error
	Update(existing, updates *management.AuthConfig) (*management.AuthConfig, error)
}

type ProviderClient struct {
	client   *management.Client
	saml     *APIClient
	session  *session.Session
	provider Provider

	Config *Config
}

// NewProviderClient constructs a SAML provider struct after it reads the provider's configuration key
func NewProviderClient(client *management.Client, samlClient *APIClient, session *session.Session, provider Provider) (*ProviderClient, error) {
	if provider.Name == "" || provider.ConfigType == "" || provider.ConfigKey == "" {
		return nil, fmt.Errorf("a SAML provider needs a name, a config type and a config key, got %+v", provider)
	}

	samlConfig := new(Config)
	config.LoadConfig(provider.ConfigKey, samlConfig)

	return &ProviderClient{
		client:   client,
		saml:     samlClient,
		session:  session,
		provider: provider,
		Config:   samlConfig,
	}, nil
}

// Provider returns the provider this client drives
func (p *ProviderClient) Provider() Provider {
	return p.provider
}

// Enable is a method of ProviderClient, writes the auth config with the given configuration values
func (p *ProviderClient) Enable() error {
	updates, err := p.newEnableInput()
	if err != nil {
		return err
	}

	if err := p.writeConfig(updates); err != nil {
		return fmt.Errorf("enabling the %s auth provider: %w", p.provider.Name, err)
	}

	p.session.RegisterCleanupFunc(p.Disable)

	return nil
}

// EnableWithAdminLogin is a method of ProviderClient, turns the provider on by having an administrator sign in
func (p *ProviderClient) EnableWithAdminLogin(username, password string) error {
	if p.saml == nil {
		return fmt.Errorf("the %s provider client has no SAML client, so it cannot complete the login "+
			"that enables the provider", p.provider.Name)
	}

	if username == "" || password == "" {
		return fmt.Errorf("enabling %s needs the credentials of an identity provider account for the "+
			"administrator, set users.admin under the %s config key", p.provider.Name, p.provider.ConfigKey)
	}

	updates, err := p.newEnableInput()
	if err != nil {
		return err
	}
	updates["enabled"] = false

	if err := p.writeConfig(updates); err != nil {
		return fmt.Errorf("writing the %s auth config before enabling it: %w", p.provider.Name, err)
	}

	warmErr := p.warmServiceProviders()

	var refused *samlext.ACSResult

	err = kwait.PollUntilContextTimeout(context.TODO(), defaults.OneSecondTimeout, defaults.OneMinuteTimeout, true,
		func(ctx context.Context) (bool, error) {
			result, err := p.enableLogin(username, password)
			if err != nil {
				return false, err
			}

			if result.Accepted {
				refused = nil

				return true, nil
			}

			refused = result

			return false, nil
		})
	if err != nil && !kwait.Interrupted(err) {
		return err
	}

	if refused != nil {
		return fmt.Errorf("Rancher did not accept the assertion that enables %s in %s, it answered %d "+
			"and redirected to %s%s%s", p.provider.Name, defaults.OneMinuteTimeout, refused.StatusCode,
			refused.Location, rejectionHint(refused), warmDescription(warmErr))
	}

	if err != nil {
		return err
	}

	p.session.RegisterCleanupFunc(p.Disable)

	enabled, err := p.client.AuthConfig.ByID(p.provider.Name)
	if err != nil {
		return fmt.Errorf("retrieving the %s auth config after enabling it: %w", p.provider.Name, err)
	}

	if !enabled.Enabled {
		return fmt.Errorf("Rancher accepted the assertion for %s but the auth config is still disabled",
			p.provider.Name)
	}

	return p.waitForServiceProviders()
}

// CaptureAssertion signs a user in and returns the assertion the identity provider issued
func (p *ProviderClient) CaptureAssertion(username, password string) (*samlext.CapturedAssertion, error) {
	userSession, err := p.saml.AuthenticateUser(p.provider, username, password)
	if err != nil {
		return nil, fmt.Errorf("signing %s in at the %s identity provider: %w", username, p.provider.Name, err)
	}

	return userSession.CaptureAssertion(p.provider)
}

func (p *ProviderClient) enableLogin(username, password string) (*samlext.ACSResult, error) {
	loginRequest, err := samlext.InitiateTestAndEnable(p.saml.noRedirectClient, p.saml.rancherURL,
		p.provider.Name, p.provider.ConfigType, p.client.Opts.TokenKey, p.saml.finalRedirectURL())
	if err != nil {
		return nil, fmt.Errorf("asking Rancher to enable %s: %w", p.provider.Name, err)
	}

	userSession, assertionDocument, err := p.saml.AuthenticateForRequest(p.provider, username, password,
		loginRequest)
	if err != nil {
		return nil, fmt.Errorf("signing the administrator in at the %s identity provider: %w", p.provider.Name, err)
	}

	assertion, err := userSession.AssertionForRequest(p.provider, loginRequest, assertionDocument)
	if err != nil {
		return nil, fmt.Errorf("collecting the assertion that enables %s: %w", p.provider.Name, err)
	}

	result, err := p.saml.SubmitAssertion(assertion, "")
	if err != nil {
		return nil, fmt.Errorf("presenting the assertion that enables %s: %w", p.provider.Name, err)
	}

	return result, nil
}

func (p *ProviderClient) warmServiceProviders() error {
	return p.settleServiceProviders(p.primeServiceProvider)
}

func (p *ProviderClient) waitForServiceProviders() error {
	return p.settleServiceProviders(nil)
}

func (p *ProviderClient) primeServiceProvider() error {
	_, err := samlext.InitiateTestAndEnable(p.saml.noRedirectClient, p.saml.rancherURL, p.provider.Name,
		p.provider.ConfigType, p.client.Opts.TokenKey, p.saml.finalRedirectURL())

	return err
}

func (p *ProviderClient) settleServiceProviders(prime func() error) error {
	certificate := p.Config.SpCert
	if certificate == "" {
		return fmt.Errorf("the config carries no service provider certificate to recognise a settled replica by")
	}

	lastErr := fmt.Errorf("the descriptor was never read")
	streak := 0

	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.OneSecondTimeout, defaults.NinetySecondTimeout,
		true, func(ctx context.Context) (bool, error) {
			if prime != nil {
				if err := prime(); err != nil {
					lastErr = err
					streak = 0

					return false, nil
				}
			}

			metadata, err := samlext.FetchServiceProviderMetadata(p.saml.noRedirectClient, p.saml.rancherURL, "",
				p.provider.Name)

			switch {
			case err != nil:
				lastErr = err
				streak = 0
			case !metadata.HasCertificate(certificate):
				lastErr = fmt.Errorf("a replica publishes a descriptor for %s carrying a different signing "+
					"certificate", p.provider.Name)
				streak = 0
			default:
				streak++
			}

			return streak >= serviceProviderWarmStreak, nil
		})
	if err != nil {
		return fmt.Errorf("spent %s waiting for every Rancher replica to serve the %s service provider: %w",
			defaults.NinetySecondTimeout, p.provider.Name, lastErr)
	}

	return nil
}

func warmDescription(warmErr error) string {
	if warmErr == nil {
		return ""
	}

	return fmt.Sprintf(". Not every Rancher replica was confirmed to have built the service provider "+
		"beforehand, which would explain a rejection: %s", warmErr)
}

func rejectionHint(result *samlext.ACSResult) string {
	if !strings.Contains(result.Location, "errorCode=403") || strings.Contains(result.Location, samlext.DashboardVerifyPath) {
		return ""
	}

	return ". A bare /login?errorCode=403 means Rancher rejected the assertion while validating it, before it " +
		"looked at the access mode or the principals, so allowed principal IDs and the access mode are not the " +
		"cause. Either the assertion answered no request the replica was tracking, which is what a login spread " +
		"across replicas looks like, or the signature did not verify against the identity provider metadata, or " +
		"the destination or audience did not match, or the assertion was outside its validity window. Rancher " +
		"names which of those it was in a debug log, so raise its log level to see it"
}

func (p *ProviderClient) writeConfig(updates map[string]any) error {
	existing, err := p.client.AuthConfig.ByID(p.provider.Name)
	if err != nil {
		return fmt.Errorf("retrieving the %s auth config: %w", p.provider.Name, err)
	}

	var result management.AuthConfig

	return p.client.Ops.DoUpdate(p.provider.ConfigType, &existing.Resource, updates, &result)
}

// Disable is a method of ProviderClient, makes a request to disable the provider
func (p *ProviderClient) Disable() error {
	existing, err := p.client.AuthConfig.ByID(p.provider.Name)
	if err != nil {
		return fmt.Errorf("retrieving the %s auth config: %w", p.provider.Name, err)
	}

	updates := map[string]any{
		"type":    p.provider.ConfigType,
		"enabled": false,
	}

	var result management.AuthConfig
	if err := p.client.Ops.DoUpdate(p.provider.ConfigType, &existing.Resource, updates, &result); err != nil {
		return fmt.Errorf("disabling the %s auth provider: %w", p.provider.Name, err)
	}

	return nil
}

// Update is a method of ProviderClient, makes an update with the given configuration values
func (p *ProviderClient) Update(existing, updates *management.AuthConfig) (*management.AuthConfig, error) {
	return p.client.AuthConfig.Update(existing, updates)
}

// UpdateAccessMode changes who may sign in through the provider
func (p *ProviderClient) UpdateAccessMode(accessMode string, allowedPrincipalIDs []string) error {
	existing, err := p.client.AuthConfig.ByID(p.provider.Name)
	if err != nil {
		return fmt.Errorf("retrieving the %s auth config: %w", p.provider.Name, err)
	}

	updates, err := p.newEnableInput()
	if err != nil {
		return err
	}
	updates["accessMode"] = accessMode
	updates["allowedPrincipalIds"] = allowedPrincipalIDs

	var result management.AuthConfig
	if err := p.client.Ops.DoUpdate(p.provider.ConfigType, &existing.Resource, updates, &result); err != nil {
		return fmt.Errorf("setting the %s access mode to %s: %w", p.provider.Name, accessMode, err)
	}

	return nil
}

func (p *ProviderClient) ensureKeyPair() error {
	if p.Config.SpCert != "" && p.Config.SpKey != "" {
		return nil
	}

	if p.Config.SpCert != "" || p.Config.SpKey != "" {
		return fmt.Errorf("only one of spCert and spKey is set under the %s config key, set both to use "+
			"your own pair or neither to have one generated", p.provider.ConfigKey)
	}

	keyPair, err := NewKeyPair(p.provider.Name)
	if err != nil {
		return err
	}

	p.Config.SpCert = keyPair.Certificate
	p.Config.SpKey = keyPair.PrivateKey

	return nil
}

func (p *ProviderClient) newEnableInput() (map[string]any, error) {
	if p.Config.RancherAPIHost == "" {
		return nil, fmt.Errorf("rancherApiHost is empty under the %s config key, the identity provider "+
			"issues assertions for this host so it must be set", p.provider.ConfigKey)
	}

	if p.Config.IDPMetadataContent == "" {
		return nil, fmt.Errorf("idpMetadataContent is empty under the %s config key, Rancher cannot trust "+
			"the identity provider without its metadata", p.provider.ConfigKey)
	}

	if err := p.ensureKeyPair(); err != nil {
		return nil, err
	}

	return map[string]any{
		"type":                p.provider.ConfigType,
		"enabled":             true,
		"accessMode":          p.Config.AccessMode,
		"allowedPrincipalIds": p.Config.AllowedPrincipalIDs,
		"rancherApiHost":      p.Config.RancherAPIHost,
		"entityID":            p.Config.EntityID,
		"displayNameField":    p.Config.DisplayNameField,
		"groupsField":         p.Config.GroupsField,
		"uidField":            p.Config.UIDField,
		"userNameField":       p.Config.UserNameField,
		"idpMetadataContent":  p.Config.IDPMetadataContent,
		"spCert":              p.Config.SpCert,
		"spKey":               p.Config.SpKey,
	}, nil
}
