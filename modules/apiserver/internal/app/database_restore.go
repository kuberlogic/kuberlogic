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
	"github.com/kuberlogic/kuberlogic/modules/apiserver/util"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
)

func (srv *Service) DatabaseRestoreHandler(params apiService.DatabaseRestoreParams, principal *models.Principal) middleware.Responder {
	service := params.HTTPRequest.Context().Value("service").(*kuberlogicv1.KuberLogicService)
	ns, name := service.Namespace, service.Name

	srv.log.Debugw("attempting to create a restore backup resource", "namespace", ns, "name", name)
	err := util.CreateBackupRestoreResource(srv.kuberlogicClient, ns, name, *params.RestoreItem.Key, *params.RestoreItem.Database)
	if err != nil {
		srv.log.Errorw("error creating a backup restore resource",
			"namespace", ns, "name", name, "error", err)
		return util.BadRequestFromError(err)
	}

	return apiService.NewDatabaseRestoreOK()
}
