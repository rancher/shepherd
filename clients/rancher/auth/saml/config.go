package saml

import management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"

type Provider struct {
	Name       string
	ConfigType string
	ConfigKey  string
}

// String stringer for the Provider
func (p Provider) String() string {
	return p.Name
}

var KeycloakSAML = Provider{
	Name:       "keycloak",
	ConfigType: management.KeyCloakConfigType,
	ConfigKey:  "keycloaksaml",
}

// Config represents the SAML authentication configuration structure
type Config struct {
	AccessMode          string   `json:"accessMode" yaml:"accessMode" default:"unrestricted"`
	AllowedPrincipalIDs []string `json:"allowedPrincipalIds" yaml:"allowedPrincipalIds"`
	RancherAPIHost      string   `json:"rancherApiHost" yaml:"rancherApiHost"`
	EntityID            string   `json:"entityID" yaml:"entityID"`
	DisplayNameField    string   `json:"displayNameField" yaml:"displayNameField" default:"givenName"`
	GroupsField         string   `json:"groupsField" yaml:"groupsField" default:"member"`
	UIDField            string   `json:"uidField" yaml:"uidField" default:"email"`
	UserNameField       string   `json:"userNameField" yaml:"userNameField" default:"email"`
	IDPMetadataContent  string   `json:"idpMetadataContent" yaml:"idpMetadataContent"`
	SpCert              string   `json:"spCert" yaml:"spCert"`
	SpKey               string   `json:"spKey" yaml:"spKey"`
	Group               string   `json:"group" yaml:"group"`
	NestedGroup         string   `json:"nestedGroup" yaml:"nestedGroup"`
	DoubleNestedGroup   string   `json:"doubleNestedGroup" yaml:"doubleNestedGroup"`
	Users               *Users   `json:"users" yaml:"users"`
}

// Users represents SAML users, used in test scenarios for validating user authentication.
type Users struct {
	Admin               *User  `json:"admin" yaml:"admin"`
	Members             []User `json:"members" yaml:"members"`
	Outsiders           []User `json:"outsiders" yaml:"outsiders"`
	NestedMembers       []User `json:"nestedMembers" yaml:"nestedMembers"`
	DoubleNestedMembers []User `json:"doubleNestedMembers" yaml:"doubleNestedMembers"`
}

// User represents a SAML user with authentication credentials, used in test scenarios for validating user authentication.
type User struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}
