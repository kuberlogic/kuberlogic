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
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/util"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
)

func (srv *Service) RestoreListHandler(params apiService.RestoreListParams, principal *models.Principal) middleware.Responder {
	service := params.HTTPRequest.Context().Value("service").(*kuberlogicv1.KuberLogicService)
	ns, name := principal.Namespace, service.Name

	srv.log.Debugw("searching for service", "namespace", ns, "name", name)
	if _, found, errGet := srv.serviceStore.GetService(name, ns, params.HTTPRequest.Context()); errGet != nil {
		srv.log.Errorw("service get error", "error", errGet.Err)
		return apiService.NewRestoreListServiceUnavailable().WithPayload(&models.Error{Message: errGet.ClientMsg})
	} else if !found {
		return util.BadRequestFromError(fmt.Errorf("%s/%s service not found", ns, name))
	}

	srv.log.Debugw("service exists", "namespace", ns, "name", name)
	restores, errRestores := srv.serviceStore.GetServiceRestores(ns, name, params.HTTPRequest.Context())
	if errRestores != nil {
		return apiService.NewRestoreListServiceUnavailable().WithPayload(&models.Error{Message: errRestores.ClientMsg})
	}

	return apiService.NewRestoreListOK().WithPayload(restores)
}
