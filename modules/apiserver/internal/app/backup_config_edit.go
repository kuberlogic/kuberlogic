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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// curl -v -H Content-Type:application/json -H "Authorization: Bearer" -X PUT localhost:8001/api/v1/services/<service-id>/backup-config -d '{"aws_access_key_id":"","aws_secret_access_key":"","bucket":"cloudmanaged","endpoint":"https://fra1.digitaloceanspaces.com","schedule":"* 1 * * *","type":"s3","enabled":false}'
func (srv *Service) BackupConfigEditHandler(params apiService.BackupConfigEditParams, principal *models.Principal) middleware.Responder {
	service := params.HTTPRequest.Context().Value("service").(*kuberlogicv1.KuberLogicService)
	ns, name := principal.Namespace, service.Name

	secretResource := util.BackupConfigModelToResource(params.BackupConfig)
	secretResource.ObjectMeta = v1.ObjectMeta{
		Name:      name,
		Namespace: ns,
	}

	srv.log.Debugw("attempting to update a backup config",
		"namespace", ns, "name", name)
	updatedResource, err := srv.clientset.CoreV1().Secrets(ns).
		Update(context.TODO(), secretResource, v1.UpdateOptions{})
	if err != nil {
		srv.log.Errorw("error updating a backup config", "error", err)
		return util.BadRequestFromError(err)
	}

	if *params.BackupConfig.Enabled {
		srv.log.Debugw("attempting to create a backup resource",
			"namespace", ns, "name", name)
		err = util.CreateBackupResource(srv.kuberlogicClient, ns, name, *params.BackupConfig.Schedule)
		if err != nil {
			srv.log.Errorw("error create a backup resource", "error", err)
			return util.BadRequestFromError(err)
		}

		srv.log.Debugw("attempting to update a backup resource",
			"namespace", ns, "name", name)
		err = util.UpdateBackupResource(srv.kuberlogicClient, ns, name, *params.BackupConfig.Schedule)
		if err != nil {
			srv.log.Errorw("error update a backup resource", "error", err)
			return util.BadRequestFromError(err)
		}
	} else {
		srv.log.Debugw("attempting to delete a backup resource",
			"namespace", ns, "name", name)
		err = util.DeleteBackupResource(srv.kuberlogicClient, ns, name)
		if err != nil {
			srv.log.Errorw("error deleting backup resource", "error", err)
			return util.BadRequestFromError(err)
		}
	}

	return &apiService.BackupConfigEditOK{
		Payload: util.BackupConfigResourceToModel(updatedResource),
	}
}
