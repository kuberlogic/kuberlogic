package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/models"
	apiAuth "github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/restapi/operations/auth"
)

func (srv *Service) LoginUserHandler(params apiAuth.LoginUserParams) middleware.Responder {
	data, err := srv.authProvider.GetAuthenticationSecret(*params.UserCredentials.Username, *params.UserCredentials.Password)
	if err != nil {
		srv.log.Errorw("error getting authentication secret for",
			"name", *params.UserCredentials.Username, "error", err)
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
