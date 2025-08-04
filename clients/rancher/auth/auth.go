package auth

import (
	"github.com/rancher/shepherd/clients/rancher/auth/openldap"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/pkg/session"
)

type Client struct {
	OLDAP *openldap.OLDAPClient
}

// NewAuth constructs the Auth Provider Struct
func NewClient(mgmt *management.Client, session *session.Session) (*Client, error) {
	oLDAP, err := openldap.NewOLDAP(mgmt, session)
	if err != nil {
		return nil, err
	}

	return &Client{
		OLDAP: oLDAP,
	}, nil
}
