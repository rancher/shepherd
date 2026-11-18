package auth

import (
	"strings"

	"github.com/rancher/shepherd/clients/rancher/auth/activedirectory"
	"github.com/rancher/shepherd/clients/rancher/auth/namedconfig"
	"github.com/rancher/shepherd/clients/rancher/auth/oidc"
	"github.com/rancher/shepherd/clients/rancher/auth/oidcprovider"
	"github.com/rancher/shepherd/clients/rancher/auth/openldap"
	"github.com/rancher/shepherd/clients/rancher/auth/saml"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/pkg/session"
)

const managementAPIPath = "/v3"

type Client struct {
	mgmt            *management.Client
	OLDAP           *openldap.OLDAPClient
	ActiveDirectory *activedirectory.Client
	OIDC            *oidc.APIClient
	SAML            *saml.APIClient
	KeycloakSAML    *saml.ProviderClient
	GenericSAML     *saml.ProviderClient
	KeycloakOIDC    *oidcprovider.ProviderClient
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

	genericSAMLClient, err := saml.NewProviderClient(mgmt, samlClient, session, saml.GenericSAML)
	if err != nil {
		return nil, err
	}

	keycloakOIDCClient, err := oidcprovider.NewProviderClient(mgmt, session, oidcprovider.KeycloakOIDC)
	if err != nil {
		return nil, err
	}

	return &Client{
		mgmt:            mgmt,
		OLDAP:           oLDAP,
		ActiveDirectory: activeDirectory,
		OIDC:            oidcClient,
		SAML:            samlClient,
		KeycloakSAML:    keycloakClient,
		GenericSAML:     genericSAMLClient,
		KeycloakOIDC:    keycloakOIDCClient,
	}, nil
}

// ListProviderTypes returns what Rancher's public auth provider types endpoint reports about the auth providers it offers
func (c *Client) ListProviderTypes() ([]map[string]any, error) {
	return namedconfig.ListProviderTypes(c.mgmt)
}
