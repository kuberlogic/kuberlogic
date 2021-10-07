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
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/util"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// curl -v -H Content-Type:application/json -H "Authorization: Bearer" -X GET localhost:8001/api/v1/services/<service-id>/backup-config
func (srv *Service) BackupConfigGetHandler(params apiService.BackupConfigGetParams, principal *models.Principal) middleware.Responder {
	service := params.HTTPRequest.Context().Value("service").(*kuberlogicv1.KuberLogicService)
	ns, name := principal.Namespace, service.Name

	srv.log.Debugw("attempting to get a backup config",
		"namespace", ns, "name", name)
	secret, err := srv.clientset.CoreV1().
		Secrets(ns).
		Get(context.TODO(), name, v1.GetOptions{})
	if errors.IsNotFound(err) {
		srv.log.Errorw("backup config does not exist",
			"namespace", ns, "name", name, "error", err)
		return &apiService.BackupConfigGetNotFound{}
	} else if err != nil {
		srv.log.Errorw("failed to get a backup config",
			"namespace", ns, "name", name, "error", err)
		return util.BadRequestFromError(err)
	}

	return &apiService.BackupConfigGetOK{
		Payload: util.BackupConfigResourceToModel(secret),
	}
}
