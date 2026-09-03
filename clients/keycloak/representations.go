package keycloak

// RealmRepresentation represents a Keycloak realm
type RealmRepresentation struct {
	ID          string `json:"id,omitempty"`
	Realm       string `json:"realm"`
	DisplayName string `json:"displayName,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

// ClientRepresentation represents a client registered in a realm
type ClientRepresentation struct {
	ID                 string                         `json:"id,omitempty"`
	ClientID           string                         `json:"clientId"`
	Name               string                         `json:"name,omitempty"`
	Description        string                         `json:"description,omitempty"`
	Protocol           string                         `json:"protocol,omitempty"`
	Enabled            *bool                          `json:"enabled,omitempty"`
	RedirectURIs       []string                       `json:"redirectUris,omitempty"`
	BaseURL            string                         `json:"baseUrl,omitempty"`
	AdminURL           string                         `json:"adminUrl,omitempty"`
	FrontchannelLogout *bool                          `json:"frontchannelLogout,omitempty"`
	Attributes         map[string]string              `json:"attributes,omitempty"`
	ProtocolMappers    []ProtocolMapperRepresentation `json:"protocolMappers,omitempty"`
}

// ProtocolMapperRepresentation represents a mapper that shapes what a client's assertion carries
type ProtocolMapperRepresentation struct {
	ID              string            `json:"id,omitempty"`
	Name            string            `json:"name"`
	Protocol        string            `json:"protocol"`
	ProtocolMapper  string            `json:"protocolMapper"`
	ConsentRequired *bool             `json:"consentRequired,omitempty"`
	Config          map[string]string `json:"config"`
}

// UserRepresentation represents an account in a realm
type UserRepresentation struct {
	ID            string                     `json:"id,omitempty"`
	Username      string                     `json:"username"`
	Email         string                     `json:"email,omitempty"`
	FirstName     string                     `json:"firstName,omitempty"`
	LastName      string                     `json:"lastName,omitempty"`
	Enabled       *bool                      `json:"enabled,omitempty"`
	EmailVerified *bool                      `json:"emailVerified,omitempty"`
	Credentials   []CredentialRepresentation `json:"credentials,omitempty"`
}

// CredentialRepresentation represents a secret held against an account
type CredentialRepresentation struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Temporary *bool  `json:"temporary,omitempty"`
}

// GroupRepresentation represents a group in a realm
type GroupRepresentation struct {
	ID        string                `json:"id,omitempty"`
	Name      string                `json:"name"`
	Path      string                `json:"path,omitempty"`
	SubGroups []GroupRepresentation `json:"subGroups,omitempty"`
}

// Pointer returns the address of a value, for the optional fields above
func Pointer[T any](value T) *T {
	return &value
}
