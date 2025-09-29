package activedirectory

const (
	ConfigurationFileKey = "activeDirectory"
)

// Config represents the Active Directory authentication configuration structure
// used for configuring AD connection parameters, user search settings,
// and group membership configuration.
type Config struct {
	Hostname       string          `json:"hostname"       yaml:"hostname"`
	IP             string          `json:"IP"             yaml:"IP"`
	Port           int             `json:"port"           yaml:"port"           default:"389"`
	ServiceAccount *ServiceAccount `json:"serviceAccount" yaml:"serviceAccount"`
	Groups         *Groups         `json:"groups"         yaml:"groups"`
	Users          *Users          `json:"users"          yaml:"users"`
	AccessMode     string          `json:"accessMode"     yaml:"accessMode"     default:"unrestricted"`
	TLS            bool            `json:"tls"            yaml:"tls"            default:"false"`
	StartTLS       bool            `json:"startTLS"       yaml:"startTLS"       default:"false"`
}

type ServiceAccount struct {
	DistinguishedName string `json:"distinguishedName" yaml:"distinguishedName"`
	Password          string `json:"password"          yaml:"password"`
}

// Groups represents Active Directory Groups, used in test scenarios for validating Groups search.
type Groups struct {
	SearchBase                   string `json:"searchBase"                    yaml:"searchBase"`
	ObjectClass                  string `json:"objectClass"                   yaml:"objectClass"                   default:"group"`
	NameAttribute                string `json:"nameAttribute"                 yaml:"nameAttribute"                 default:"name"`
	SearchAttribute              string `json:"searchAttribute"               yaml:"searchAttribute"               default:"sAMAccountName"`
	SearchFilter                 string `json:"searchFilter"                  yaml:"searchFilter"`
	MemberMappingAttribute       string `json:"memberMappingAttribute"        yaml:"memberMappingAttribute"        default:"member"`
	MemberUserAttribute          string `json:"memberUserAttribute"           yaml:"memberUserAttribute"           default:"distinguishedName"`
	DNAttribute                  string `json:"dnAttribute"                   yaml:"dnAttribute"                   default:"distinguishedName"`
	NestedGroupMembershipEnabled bool   `json:"nestedGroupMembershipEnabled"  yaml:"nestedGroupMembershipEnabled"  default:"true"`
}

// Users represents Active Directory users, used in test scenarios for validating users search.
type Users struct {
	Admin             *User  `json:"admin"                yaml:"admin"`
	SearchBase        string `json:"searchBase"           yaml:"searchBase"`
	ObjectClass       string `json:"objectClass"          yaml:"objectClass"          default:"user"`
	UsernameAttribute string `json:"usernameAttribute"    yaml:"usernameAttribute"    default:"name"`
	LoginAttribute    string `json:"loginAttribute"       yaml:"loginAttribute"       default:"sAMAccountName"`
	MemberAttribute   string `json:"memberAttribute"      yaml:"memberAttribute"      default:"memberOf"`
	SearchAttribute   string `json:"searchAttribute"      yaml:"searchAttribute"      default:"sAMAccountName|sn|givenName"`
	SearchFilter      string `json:"searchFilter"         yaml:"searchFilter"`
	LoginFilter       string `json:"loginFilter"          yaml:"loginFilter"`
	EnabledAttribute  string `json:"enabledAttribute"     yaml:"enabledAttribute"     default:"userAccountControl"`
	DisabledBitMask   int64  `json:"disabledBitMask"      yaml:"disabledBitMask"      default:"2"`
}

// User represents an Active Directory user with authentication credentials, used in test scenarios for validating user authentication.
type User struct {
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
}
