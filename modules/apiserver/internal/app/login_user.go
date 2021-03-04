package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiAuth "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/auth"
)

func (srv *Service) LoginUserHandler(params apiAuth.LoginUserParams) middleware.Responder {
	data, err := srv.authProvider.GetAuthenticationSecret(*params.UserCredentials.Username, *params.UserCredentials.Password)
	if err != nil {
		srv.log.Errorf("error getting authentication secret for %s: %s", *params.UserCredentials.Username, err.Error())
		return apiAuth.NewLoginUserUnauthorized()
	} else {
		a := apiAuth.NewLoginUserOK()
		d := models.AccessTokenResponse{
			AccessToken:  data,
			ExpiresIn:    0,
			RefreshToken: "",
			TokenType:    "",
		}
		a.Payload = &d
		return a
	}
}
