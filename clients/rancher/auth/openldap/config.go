package openldap

const (
	ConfigurationFileKey = "openLDAP"
)

// Config represents the OpenLDAP authentication configuration structure
// used for configuring LDAP connection parameters, user search settings,
// and group membership configuration.
type Config struct {
	Hostname       string          `json:"hostname"       yaml:"hostname"`
	IP             string          `json:"IP"             yaml:"IP"`
	ServiceAccount *ServiceAccount `json:"serviceAccount" yaml:"serviceAccount"`
	Groups         *Groups         `json:"groups"          yaml:"groups"`
	Users          *Users          `json:"users"          yaml:"users"`
	AccessMode     string          `json:"accessMode"     yaml:"accessMode"     default:"unrestricted"`
}

type ServiceAccount struct {
	DistinguishedName string `json:"distinguishedName" yaml:"distinguishedName"`
	Password          string `json:"password"          yaml:"password"`
}

// Users represents  LDAP Groups, used in test scenarios for validating Groups search.
type Groups struct {
	ObjectClass                  string `json:"objectClass"            yaml:"objectClass"`
	MemberMappingAttribute       string `json:"memberMappingAttribute" yaml:"memberMappingAttribute"`
	NestedGroupMembershipEnabled bool   `json:"nestedGroupMembershipEnabled,omitempty" yaml:"nestedGroupMembershipEnabled,omitempty"`
	SearchDirectGroupMemberships bool   `json:"searchDirectGroupMemberships,omitempty" yaml:"searchDirectGroupMemberships,omitempty"`
	SearchBase                   string `json:"searchBase" yaml:"searchBase"`
}

// Users represents  LDAP users, used in test scenarios for validating users search.

type Users struct {
	Admin      *User  `json:"admin"      yaml:"admin"`
	SearchBase string `json:"searchBase" yaml:"searchBase"`
}

// User represents an LDAP user with authentication credentials, used in test scenarios for validating user authentication.
type User struct {
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
}
