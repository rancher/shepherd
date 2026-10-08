package oidcprovider

import management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"

// Provider names an OpenID Connect authentication provider Rancher can be pointed at
type Provider struct {
	Name       string
	ConfigType string
	PublicType string
	ConfigKey  string
}

// String stringer for the Provider
func (p Provider) String() string {
	return p.Name
}

var KeycloakOIDC = Provider{
	Name:       "keycloakoidc",
	ConfigType: management.KeyCloakOIDCConfigType,
	PublicType: "keyCloakOIDCProvider",
	ConfigKey:  "keycloakoidc",
}

// Config represents the OpenID Connect authentication configuration structure
type Config struct {
	AccessMode                string   `json:"accessMode" yaml:"accessMode" default:"unrestricted"`
	AllowedPrincipalIDs       []string `json:"allowedPrincipalIds" yaml:"allowedPrincipalIds"`
	ClientID                  string   `json:"clientId" yaml:"clientId"`
	ClientSecret              string   `json:"clientSecret" yaml:"clientSecret"`
	Issuer                    string   `json:"issuer" yaml:"issuer"`
	AuthEndpoint              string   `json:"authEndpoint" yaml:"authEndpoint"`
	TokenEndpoint             string   `json:"tokenEndpoint" yaml:"tokenEndpoint"`
	UserInfoEndpoint          string   `json:"userInfoEndpoint" yaml:"userInfoEndpoint"`
	JWKSUrl                   string   `json:"jwksUrl" yaml:"jwksUrl"`
	EndSessionEndpoint        string   `json:"endSessionEndpoint" yaml:"endSessionEndpoint"`
	RancherURL                string   `json:"rancherUrl" yaml:"rancherUrl"`
	Certificate               string   `json:"certificate" yaml:"certificate"`
	PrivateKey                string   `json:"privateKey" yaml:"privateKey"`
	Scopes                    string   `json:"scope" yaml:"scope" default:"openid profile email"`
	GroupsClaim               string   `json:"groupsClaim" yaml:"groupsClaim" default:"groups"`
	PKCEMethod                string   `json:"pkceMethod" yaml:"pkceMethod"`
	GroupSearchEnabled        *bool    `json:"groupSearchEnabled" yaml:"groupSearchEnabled"`
	ClientAuthenticatedSearch bool     `json:"clientAuthenticatedSearch" yaml:"clientAuthenticatedSearch"`
	Group                     string   `json:"group" yaml:"group"`
	NestedGroup               string   `json:"nestedGroup" yaml:"nestedGroup"`
	DoubleNestedGroup         string   `json:"doubleNestedGroup" yaml:"doubleNestedGroup"`
	Users                     *Users   `json:"users" yaml:"users"`
}

// Users represents the identity provider accounts a run signs in as
type Users struct {
	Admin               *User  `json:"admin" yaml:"admin"`
	Members             []User `json:"members" yaml:"members"`
	Outsiders           []User `json:"outsiders" yaml:"outsiders"`
	NestedMembers       []User `json:"nestedMembers" yaml:"nestedMembers"`
	DoubleNestedMembers []User `json:"doubleNestedMembers" yaml:"doubleNestedMembers"`
}

// User represents an identity provider account with the credentials needed to sign it in
type User struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}
