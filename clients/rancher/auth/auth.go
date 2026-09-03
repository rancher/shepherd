package auth

import (
	"strings"

	"github.com/rancher/shepherd/clients/rancher/auth/activedirectory"
	"github.com/rancher/shepherd/clients/rancher/auth/oidc"
	"github.com/rancher/shepherd/clients/rancher/auth/openldap"
	"github.com/rancher/shepherd/clients/rancher/auth/saml"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/pkg/session"
)

const managementAPIPath = "/v3"

type Client struct {
	OLDAP           *openldap.OLDAPClient
	ActiveDirectory *activedirectory.Client
	OIDC            *oidc.APIClient
	SAML            *saml.APIClient
	KeycloakSAML    *saml.ProviderClient
}

// NewClient constructs the Auth Provider Struct
func NewClient(mgmt *management.Client, session *session.Session) (*Client, error) {
	oLDAP, err := openldap.NewOLDAP(mgmt, session)
	if err != nil {
		return nil, err
	}

	activeDirectory, err := activedirectory.NewActiveDirectory(mgmt, session)
	if err != nil {
		return nil, err
	}

	oidcClient, err := oidc.NewAPIClient(strings.TrimSuffix(mgmt.Opts.URL, managementAPIPath))
	if err != nil {
		return nil, err
	}

	samlClient, err := saml.NewAPIClient(strings.TrimSuffix(mgmt.Opts.URL, managementAPIPath))
	if err != nil {
		return nil, err
	}

	keycloakClient, err := saml.NewProviderClient(mgmt, samlClient, session, saml.KeycloakSAML)
	if err != nil {
		return nil, err
	}

	return &Client{
		OLDAP:           oLDAP,
		ActiveDirectory: activeDirectory,
		OIDC:            oidcClient,
		SAML:            samlClient,
		KeycloakSAML:    keycloakClient,
	}, nil
}
