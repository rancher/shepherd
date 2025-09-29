package auth

import (
	"github.com/rancher/shepherd/clients/rancher/auth/activedirectory"
	"github.com/rancher/shepherd/clients/rancher/auth/openldap"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/pkg/session"
)

type Client struct {
	OLDAP           *openldap.OLDAPClient
	ActiveDirectory *activedirectory.Client
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

	return &Client{
		OLDAP:           oLDAP,
		ActiveDirectory: activeDirectory,
	}, nil
}
