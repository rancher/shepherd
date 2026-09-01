package keycloak

// Config corresponds to the YAML structure for Keycloak SAML
type Config struct {
	AccessMode          string   `json:"accessMode" yaml:"accessMode"`
	AllowedPrincipalIDs []string `json:"allowedPrincipalIDs" yaml:"allowedPrincipalIDs"`
	RancherApiHost      string   `json:"rancherApiHost" yaml:"rancherApiHost"`
	DisplayNameField    string   `json:"displayNameField" yaml:"displayNameField"`
	GroupsField         string   `json:"groupsField" yaml:"groupsField"`
	UIDField            string   `json:"uidField" yaml:"uidField"`
	UserNameField       string   `json:"userNameField" yaml:"userNameField"`
	IDPMetadataContent  string   `json:"idpMetadataContent" yaml:"idpMetadataContent"`
	SpCert              string   `json:"spCert" yaml:"spCert"`
	SpKey               string   `json:"spKey" yaml:"spKey"`
	Users               *Users   `json:"users" yaml:"users"`
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
