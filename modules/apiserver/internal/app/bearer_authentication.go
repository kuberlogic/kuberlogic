package app

import (
	"github.com/go-openapi/errors"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
)

func (srv *Service) BearerAuthentication(token string) (*models.Principal, error) {
	username, bearerToken, err := srv.authProvider.Authenticate(token)
	if err != nil {
		return nil, errors.Unauthenticated("authentication failed: " + err.Error())
	}
	p := &models.Principal{
		Username: username,
		Token:    bearerToken,
	}
	return p, nil
}
