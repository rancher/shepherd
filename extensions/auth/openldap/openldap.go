package openldap

import (
	"fmt"

	management "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
)

type OLDAPOperations interface {
	Enable() error
	Disable() error
}

const (
	resourceType = "openldap"
	schemaType   = "openLdapConfigs"
)

type OLDAP struct {
	Config *Config

	session *session.Session
	client  *rancher.Client
}

// NewOLDAP constructs OLDAP struct after it reads Open LDAP from the configuration file
func NewOLDAP(client *rancher.Client, ts *session.Session) (*OLDAP, error) {
	ldapConfig := new(Config)
	config.LoadConfig(ConfigurationFileKey, ldapConfig)

	return &OLDAP{
		Config:  ldapConfig,
		session: ts,
		client:  client,
	}, nil
}

// Enable is a method of OLDAP, makes a request to the action with the given
// configuration values
func (o *OLDAP) Enable() error {
	var jsonResp map[string]interface{}

	url := o.newActionURL("testAndApply")
	enableActionInput, err := o.newEnableInputFromConfig()
	if err != nil {
		return err
	}

	err = o.client.Management.Ops.DoModify("POST", url, enableActionInput, &jsonResp)
	if err != nil {
		return err
	}

	o.session.RegisterCleanupFunc(func() error {
		return o.Disable()
	})

	return nil
}

// Disable is a method of OLDAP, makes a request to disable Open LDAP
func (o *OLDAP) Disable() error {
	var jsonResp map[string]any

	url := o.newActionURL("disable")
	disableActionInput := o.newDisableInput()

	return o.client.Management.Ops.DoModify("POST", url, &disableActionInput, &jsonResp)
}

func (o *OLDAP) newActionURL(action string) string {
	protocol := "https"

	if *o.client.RancherConfig.Insecure {
		protocol = "http"
	}

	return fmt.Sprintf("%v://%v/v3/%v/%v?action=%v", protocol, o.client.RancherConfig.Host, schemaType, resourceType, action)
}

func (o *OLDAP) newEnableInputFromConfig() (*management.LdapTestAndApplyInput, error) {
	var resource management.LdapTestAndApplyInput

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

func (o *OLDAP) newDisableInput() []byte {
	return []byte(`{"action": "disable"}`)
}
