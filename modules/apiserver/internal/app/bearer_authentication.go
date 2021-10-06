package app

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/go-openapi/errors"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/models"
)

func (srv *Service) BearerAuthentication(token string) (*models.Principal, error) {
	email, bearerToken, err := srv.authProvider.Authenticate(token)
	if err != nil {
		return nil, errors.Unauthenticated("authentication failed: " + err.Error())
	}
	p := &models.Principal{
		Email: email,
		Token: bearerToken,
		Namespace: func(string) string {
			// namespace should be DNS compliant, less than 63 chars string
			// MD5 hashing matches all of these requirements
			h := md5.Sum([]byte(email))
			return hex.EncodeToString(h[:])
		}(email),
	}
	return p, nil
}
