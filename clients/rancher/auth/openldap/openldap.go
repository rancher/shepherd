package openldap

import (
	"fmt"

	apisv3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"

	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
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
	client  *management.Client
	session *session.Session

	Config *Config
}

// NewOLDAP constructs OLDAP struct after it reads Open LDAP auth from the configuration file
func NewOLDAP(client *management.Client, session *session.Session) (*OLDAPClient, error) {
	ldapConfig := new(Config)
	config.LoadConfig(ConfigurationFileKey, ldapConfig)

	return &OLDAPClient{
		client:  client,
		session: session,
		Config:  ldapConfig,
	}, nil
}

// Enable is a method of OLDAP, makes a request to the action with the given
// configuration values
func (o *OLDAPClient) Enable() error {
	var jsonResp map[string]interface{}

	url := o.newActionURL("testAndApply")
	enableActionInput, err := o.newEnableInputFromConfig()
	if err != nil {
		return err
	}

	err = o.client.Ops.DoModify("POST", url, enableActionInput, &jsonResp)
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

// Disable is a method of OLDAP, makes a request to disable Open LDAP auth provider
func (o *OLDAPClient) Disable() error {
	var jsonResp map[string]any

	url := o.newActionURL("disable")
	disableActionInput := o.newDisableInput()

	return o.client.Ops.DoModify("POST", url, &disableActionInput, &jsonResp)
}

func (o *OLDAPClient) newActionURL(action string) string {
	return fmt.Sprintf(
		"%v/%v/%v?action=%v",
		o.client.Opts.URL,
		schemaType,
		resourceType,
		action,
	)
}

func (o *OLDAPClient) newEnableInputFromConfig() (*apisv3.LdapTestAndApplyInput, error) {
	var resource apisv3.LdapTestAndApplyInput

	var server string
	if o.Config.Hostname == "" && o.Config.IP == "" {
		return nil, fmt.Errorf("open LDAP Hostname and IP are empty, please provide one of them")
	}
	server = o.Config.Hostname
	if server == "" {
		server = o.Config.IP
	}

	resource.Enabled = true
	resource.AccessMode = o.Config.AccessMode

	resource.UserSearchBase = o.Config.Users.SearchBase

	if o.Config.Users.Admin.Username == "" || o.Config.Users.Admin.Password == "" {
		return nil, fmt.Errorf("admin username or password are empty, please provide them")
	}

	resource.Username = o.Config.Users.Admin.Username
	resource.Password = o.Config.Users.Admin.Password

	resource.Servers = []string{server}

	resource.ServiceAccountDistinguishedName = o.Config.ServiceAccount.DistinguishedName
	resource.ServiceAccountPassword = o.Config.ServiceAccount.Password

	resource.GroupMemberUserAttribute = o.Config.Group.MemberMappingAttribute
	resource.GroupObjectClass = o.Config.Group.ObjectClass

	return &resource, nil
}

func (o *OLDAPClient) newDisableInput() []byte {
	return []byte(`{"action": "disable"}`)
}
