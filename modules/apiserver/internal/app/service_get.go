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
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
)

func (srv *Service) ServiceGetHandler(params apiService.ServiceGetParams, principal *models.Principal) middleware.Responder {
	kuberlogicService := params.HTTPRequest.Context().Value("service").(*kuberlogicv1.KuberLogicService)
	ns, name := principal.Namespace, kuberlogicService.Name
	// TODO: need to use kuberLogicToService directly

	service, found, errGet := srv.serviceStore.GetService(name, ns, params.HTTPRequest.Context())
	if errGet != nil {
		srv.log.Errorw("service get error", "error", errGet.Err)
		return apiService.NewServiceGetServiceUnavailable().WithPayload(&models.Error{Message: errGet.ClientMsg})
	}
	if !found {
		return apiService.NewServiceGetNotFound()
	}

	return apiService.NewServiceGetOK().WithPayload(service)
}
