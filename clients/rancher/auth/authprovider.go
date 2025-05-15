package auth

type Provider string

const (
	LocalAuth    Provider = "local"
	OpenLDAPAuth Provider = "openLdap"
)

// String stringer for the AuthProvider
func (a Provider) String() string {
	return string(a)
}
