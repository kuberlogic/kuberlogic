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

func (srv *Service) BackupConfigDeleteHandler(params apiService.BackupConfigDeleteParams, principal *models.Principal) middleware.Responder {
	service := params.HTTPRequest.Context().Value("service").(*kuberlogicv1.KuberLogicService)
	ns, name := principal.Namespace, service.Name

	srv.log.Debugw("attempting to delete a backup config", "namespace", ns, "name", name)
	err := srv.clientset.CoreV1().Secrets(ns).
		Delete(context.TODO(), name, v1.DeleteOptions{})
	if errors.IsNotFound(err) {
		srv.log.Errorw("backup config does not exist",
			"namespace", ns, "name", name, "error", err)
		return &apiService.BackupConfigDeleteNotFound{}
	}
	if err != nil {
		srv.log.Errorw("error deleting backup config", "error", err)
		return util.BadRequestFromError(err)
	}

	srv.log.Debugw("attempting to delete a backup resource",
		"namespace", ns, "name", name)
	err = util.DeleteBackupResource(srv.kuberlogicClient, ns, name)
	if err != nil {
		srv.log.Errorw("error deleting backup resource", "error", err)
		return util.BadRequestFromError(err)
	}

	return &apiService.BackupConfigDeleteOK{}
}
