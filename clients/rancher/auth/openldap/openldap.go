package openldap

import (
	"fmt"

	"github.com/rancher/shepherd/clients/rancher/auth/namedconfig"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/pkg/clientbase"
	"github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
)

type OLDAPOperations interface {
	Enable() error
	Disable() error
	Update(existing, updates *management.AuthConfig) (*management.AuthConfig, error)
}

const (
	resourceType = "openldap"
	schemaType   = "openLdapConfigs"
)

type OLDAPClient struct {
	client      *management.Client
	session     *session.Session
	configName  string
	displayName string

	Config *Config
}

// NewOLDAP constructs OLDAP struct after it reads Open LDAP from the configuration file
func NewOLDAP(client *management.Client, session *session.Session) (*OLDAPClient, error) {
	ldapConfig := new(Config)
	config.LoadConfig(ConfigurationFileKey, ldapConfig)

	return &OLDAPClient{
		client:  client,
		session: session,
		Config:  ldapConfig,
	}, nil
}

// ForConfig returns a client that drives the auth config with the given name instead of the provider's default one.
// A nil config keeps the configuration this client was built with.
func (o *OLDAPClient) ForConfig(configName, displayName string, config *Config) *OLDAPClient {
	named := *o
	named.configName = configName
	named.displayName = displayName

	if config != nil {
		named.Config = config
	}

	return &named
}

// ConfigName returns the name of the auth config this client drives, empty for the provider's default one
func (o *OLDAPClient) ConfigName() string {
	return o.configName
}

// Enable is a method of OLDAP, makes a request to the action with the given
// configuration values
func (o *OLDAPClient) Enable() error {
	var jsonResp map[string]interface{}

	if err := namedconfig.Ensure(o.session, o.client, management.OpenLdapConfigType, o.configName, o.displayName); err != nil {
		return err
	}

	url := o.newActionURL("testAndApply")
	enableActionInput, err := o.newEnableInputFromConfig()
	if err != nil {
		return err
	}

	body, err := namedconfig.Body(enableActionInput, o.configName)
	if err != nil {
		return err
	}

	err = o.client.Ops.DoModify("POST", url, body, &jsonResp)
	if err != nil {
		return err
	}

	o.session.RegisterCleanupFunc(func() error {
		return o.Disable()
	})

	return nil
}

// Update is a method of OLDAP, makes an update with the given configuration values
func (o *OLDAPClient) Update(
	existing, updates *management.AuthConfig,
) (*management.AuthConfig, error) {
	return o.client.AuthConfig.Update(existing, updates)
}

// Disable is a method of OLDAP, makes a request to disable Open LDAP.
// A named config that is already disabled or gone is left alone, since Rancher rejects the action on it.
func (o *OLDAPClient) Disable() error {
	var jsonResp map[string]any

	if o.configName != "" {
		existing, err := o.client.AuthConfig.ByID(o.configName)
		if clientbase.IsNotFound(err) || (err == nil && !existing.Enabled) {
			return nil
		}
	}

	url := o.newActionURL("disable")
	disableActionInput := o.newDisableInput()

	return o.client.Ops.DoModify("POST", url, disableActionInput, &jsonResp)
}

// Delete removes the named auth config this client drives. The provider's default config cannot be deleted.
func (o *OLDAPClient) Delete() error {
	if o.configName == "" {
		return fmt.Errorf("the default %s auth config cannot be deleted", resourceType)
	}

	return namedconfig.Delete(o.client, o.configName)
}

func (o *OLDAPClient) newActionURL(action string) string {
	name := resourceType
	if o.configName != "" {
		name = o.configName
	}

	return fmt.Sprintf(
		"%v/%v/%v?action=%v",
		o.client.Opts.URL,
		schemaType,
		name,
		action,
	)
}

func (o *OLDAPClient) newEnableInputFromConfig() (*management.OpenLdapTestAndApplyInput, error) {
	var input management.OpenLdapTestAndApplyInput
	var ldapConfig management.LdapConfig

	var server string
	if o.Config.Hostname == "" && o.Config.IP == "" {
		return nil, fmt.Errorf("open LDAP Hostname and IP are empty, please provide one of them")
	}
	server = o.Config.Hostname
	if server == "" {
		server = o.Config.IP
	}

	ldapConfig.Enabled = true
	ldapConfig.AccessMode = o.Config.AccessMode
	ldapConfig.UserSearchBase = o.Config.Users.SearchBase
	ldapConfig.GroupSearchBase = o.Config.Groups.SearchBase
	ldapConfig.Servers = []string{server}
	if o.Config.Port > 0 {
		ldapConfig.Port = int64(o.Config.Port)
	}
	ldapConfig.ServiceAccountDistinguishedName = o.Config.ServiceAccount.DistinguishedName
	ldapConfig.ServiceAccountPassword = o.Config.ServiceAccount.Password
	ldapConfig.GroupMemberUserAttribute = o.Config.Groups.MemberMappingAttribute
	ldapConfig.NestedGroupMembershipEnabled = o.Config.Groups.NestedGroupMembershipEnabled
	ldapConfig.GroupObjectClass = o.Config.Groups.ObjectClass

	if o.Config.Users.Admin.Username == "" || o.Config.Users.Admin.Password == "" {
		return nil, fmt.Errorf("admin username or password are empty, please provide them")
	}

	input.LdapConfig = &ldapConfig
	input.Username = o.Config.Users.Admin.Username
	input.Password = o.Config.Users.Admin.Password

	return &input, nil
}

func (o *OLDAPClient) newDisableInput() map[string]string {
	if o.configName != "" {
		return map[string]string{"configName": o.configName}
	}

	return map[string]string{"configName": resourceType}
}
