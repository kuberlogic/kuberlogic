/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
