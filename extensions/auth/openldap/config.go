package openldap

const (
	ConfigurationFileKey = "openLDAP"
)

type Config struct {
	Hostname       string          `json:"hostname" yaml:"hostname"`
	IP             string          `json:"IP" yaml:"IP"`
	ServiceAccount *ServiceAccount `json:"serviceAccount" yaml:"serviceAccount"`
	Group          *Group          `json:"group" yaml:"group"`
	Users          *Users          `json:"users" yaml:"users"`
	AccessMode     string          `json:"accessMode" yaml:"accessMode" default:"unrestricted"`
}

type ServiceAccount struct {
	DistinguishedName string `json:"distinguishedName" yaml:"distinguishedName"`
	Password          string `json:"password" yaml:"password"`
}

type Group struct {
	ObjectClass                  string `json:"objectClass"            yaml:"objectClass"`
	MemberMappingAttribute       string `json:"memberMappingAttribute" yaml:"memberMappingAttribute"`
	NestedGroupMembershipEnabled bool   `json:"nestedGroupMembershipEnabled,omitempty" yaml:"nestedGroupMembershipEnabled,omitempty"`
	SearchDirectGroupMemberships bool   `json:"searchDirectGroupMemberships,omitempty" yaml:"searchDirectGroupMemberships,omitempty"`
}

type User struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

type Users struct {
	Admin      *User           `json:"admin" yaml:"admin"`
	SearchBase string          `json:"searchBase" yaml:"searchBase"`
	Others     map[string]User `json:"others" yaml:"others"`
}
