package activedirectory

import (
	"fmt"

	apisv3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
)

type ActiveDirectoryOperations interface {
	Enable() error
	Disable() error
	Update(existing, updates *management.AuthConfig) (*management.AuthConfig, error)
}

const (
	resourceType = "activedirectory"
	schemaType   = "activeDirectoryConfigs"
)

type ActiveDirectoryClient struct {
	client  *management.Client
	session *session.Session
	Config  *Config
}

// NewActiveDirectory constructs ActiveDirectory struct after it reads Active Directory from the configuration file
func NewActiveDirectory(client *management.Client, session *session.Session) (*ActiveDirectoryClient, error) {
	adConfig := new(Config)
	config.LoadConfig(ConfigurationFileKey, adConfig)

	return &ActiveDirectoryClient{
		client:  client,
		session: session,
		Config:  adConfig,
	}, nil
}

// Enable is a method of ActiveDirectory, makes a request to the action with the given
// configuration values
func (a *ActiveDirectoryClient) Enable() error {
	var jsonResp map[string]interface{}
	url := a.newActionURL("testAndApply")

	enableActionInput, err := a.newEnableInputFromConfig()
	if err != nil {
		return err
	}

	err = a.client.Ops.DoModify("POST", url, enableActionInput, &jsonResp)
	if err != nil {
		return err
	}

	a.session.RegisterCleanupFunc(func() error {
		return a.Disable()
	})

	return nil
}

// Update is a method of ActiveDirectory, makes an update with the given configuration values
func (a *ActiveDirectoryClient) Update(
	existing, updates *management.AuthConfig,
) (*management.AuthConfig, error) {
	return a.client.AuthConfig.Update(existing, updates)
}

// Disable is a method of ActiveDirectory, makes a request to disable Active Directory
func (a *ActiveDirectoryClient) Disable() error {
	var jsonResp map[string]any
	url := a.newActionURL("disable")
	disableActionInput := a.newDisableInput()

	return a.client.Ops.DoModify("POST", url, &disableActionInput, &jsonResp)
}

func (a *ActiveDirectoryClient) newActionURL(action string) string {
	return fmt.Sprintf(
		"%v/%v/%v?action=%v",
		a.client.Opts.URL,
		schemaType,
		resourceType,
		action,
	)
}

func (a *ActiveDirectoryClient) newEnableInputFromConfig() (*apisv3.ActiveDirectoryTestAndApplyInput, error) {
	var server string

	if a.Config.Hostname == "" && a.Config.IP == "" {
		return nil, fmt.Errorf("active Directory Hostname and IP are empty, please provide one of them")
	}

	server = a.Config.Hostname
	if server == "" {
		server = a.Config.IP
	}

	if a.Config.Users.Admin.Username == "" || a.Config.Users.Admin.Password == "" {
		return nil, fmt.Errorf("admin username or password are empty, please provide them")
	}

	// Create the nested ActiveDirectoryConfig
	adConfig := &apisv3.ActiveDirectoryConfig{
		AuthConfig: apisv3.AuthConfig{
			AccessMode: a.Config.AccessMode,
		},
		Servers:                      []string{server},
		Port:                         int64(a.Config.Port),
		TLS:                          a.Config.TLS,
		StartTLS:                     a.Config.StartTLS,
		ServiceAccountUsername:       a.Config.ServiceAccount.DistinguishedName,
		ServiceAccountPassword:       a.Config.ServiceAccount.Password,
		UserSearchBase:               a.Config.Users.SearchBase,
		UserObjectClass:              a.Config.Users.ObjectClass,
		UserNameAttribute:            a.Config.Users.UsernameAttribute,
		UserLoginAttribute:           a.Config.Users.LoginAttribute,
		UserSearchAttribute:          a.Config.Users.SearchAttribute,
		UserSearchFilter:             a.Config.Users.SearchFilter,
		UserEnabledAttribute:         a.Config.Users.EnabledAttribute,
		UserDisabledBitMask:          a.Config.Users.DisabledBitMask,
		GroupSearchBase:              a.Config.Groups.SearchBase,
		GroupObjectClass:             a.Config.Groups.ObjectClass,
		GroupNameAttribute:           a.Config.Groups.NameAttribute,
		GroupSearchAttribute:         a.Config.Groups.SearchAttribute,
		GroupSearchFilter:            a.Config.Groups.SearchFilter,
		GroupMemberUserAttribute:     a.Config.Groups.MemberUserAttribute,
		GroupMemberMappingAttribute:  a.Config.Groups.MemberMappingAttribute,
		GroupDNAttribute:             a.Config.Groups.DNAttribute,
		NestedGroupMembershipEnabled: &a.Config.Groups.NestedGroupMembershipEnabled,
	}

	// Wrap it in ActiveDirectoryTestAndApplyInput with test credentials
	testAndApplyInput := &apisv3.ActiveDirectoryTestAndApplyInput{
		ActiveDirectoryConfig: *adConfig,
		Username:              a.Config.Users.Admin.Username,
		Password:              a.Config.Users.Admin.Password,
		Enabled:               true,
	}

	return testAndApplyInput, nil
}

func (a *ActiveDirectoryClient) newDisableInput() []byte {
	return []byte(`{"action": "disable"}`)
}
