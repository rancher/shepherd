package openldap

const (
	ConfigurationFileKey = "openLDAP"
)

type Config struct {
	Hostname       string          `json:"hostname"       yaml:"hostname"`
	IP             string          `json:"IP"             yaml:"IP"`
	ServiceAccount *ServiceAccount `json:"serviceAccount" yaml:"serviceAccount"`
	Group          *Group          `json:"group"          yaml:"group"`
	Users          *Users          `json:"users"          yaml:"users"`
	AccessMode     string          `json:"accessMode"     yaml:"accessMode"     default:"unrestricted"`
}

type ServiceAccount struct {
	DistinguishedName string `json:"distinguishedName" yaml:"distinguishedName"`
	Password          string `json:"password"          yaml:"password"`
}

type Group struct {
	ObjectClass            string `json:"objectClass"            yaml:"objectClass"`
	MemberMappingAttribute string `json:"memberMappingAttribute" yaml:"memberMappingAttribute"`
}

type Users struct {
	Admin      *User  `json:"admin"      yaml:"admin"`
	SearchBase string `json:"searchBase" yaml:"searchBase"`
}

type User struct {
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
}
