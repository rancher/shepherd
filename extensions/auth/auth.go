package auth

import (
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/auth/openldap"
	"github.com/rancher/shepherd/pkg/session"
)

type Auth struct {
	OLDAP *openldap.OLDAP
}

// NewAuth constructs the Auth Provider Struct
func NewAuth(client *rancher.Client, session *session.Session) (*Auth, error) {
	oLDAP, err := openldap.NewOLDAP(client, session)
	if err != nil {
		return nil, err
	}

	return &Auth{
		OLDAP: oLDAP,
	}, nil
}
