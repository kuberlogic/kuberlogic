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
	apiService "github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/restapi/operations/service"
)

func (srv *Service) ServiceEditHandler(params apiService.ServiceEditParams, principal *models.Principal) middleware.Responder {
	service, errUpdate := srv.serviceStore.UpdateService(params.ServiceItem, principal, params.HTTPRequest.Context())
	if errUpdate != nil {
		srv.log.Errorw("error updating service", "error", errUpdate.Err)
		if errUpdate.Client {
			return apiService.NewServiceEditBadRequest().WithPayload(&models.Error{Message: errUpdate.ClientMsg})
		} else {
			return apiService.NewServiceEditServiceUnavailable().WithPayload(&models.Error{Message: errUpdate.ClientMsg})
		}
	}
	return apiService.NewServiceEditOK().WithPayload(service)
}
