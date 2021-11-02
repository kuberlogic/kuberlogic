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
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// curl -v -H Content-Type:application/json -H "Authorization: Bearer" -X POST localhost:8001/api/v1/services/<service-id>/backup-config -d '{"aws_access_key_id":"SJ3MEX4WE7G2A5JLHJQC","aws_secret_access_key":"hTXfI4Gbv0SPSWGhnWQrINg6TPcWCCvLcB2DRFmp+Ok","bucket":"cloudmanaged","endpoint":"https://fra1.digitaloceanspaces.com","schedule":"* 1 * * *","type":"s3","enabled":false}'
func (srv *Service) BackupConfigCreateHandler(params apiService.BackupConfigCreateParams, principal *models.Principal) middleware.Responder {
	service := params.HTTPRequest.Context().Value("service").(*kuberlogicv1.KuberLogicService)
	ns, name := principal.Namespace, service.Name

	// Create secret
	secretResource := util.BackupConfigModelToResource(params.BackupConfig)
	secretResource.ObjectMeta = v1.ObjectMeta{
		Name:      name,
		Namespace: ns,
	}

	srv.log.Debugw("attempting to create a backup config", "namespace", ns, "name", name)
	_, err := srv.clientset.CoreV1().
		Secrets(ns).
		Create(context.TODO(), secretResource, v1.CreateOptions{})
	if err != nil {
		newErr := errors.Wrap(err, "failed to create a backup config")
		srv.log.Errorw(newErr.Error(),
			"namespace", ns, "name", name)
		return util.BadRequestFromError(newErr)
	}

	if *params.BackupConfig.Enabled {
		srv.log.Debugw("attempting to create a backup resource",
			"namespace", ns, "name", name)
		err = util.CreateBackupResource(srv.kuberlogicClient, ns, name, *params.BackupConfig.Schedule)
		if err != nil {
			newErr := errors.Wrap(err, "error creating a backup resource")
			srv.log.Errorw(newErr.Error(), "namespace", ns, "name", name)
			return util.BadRequestFromError(newErr)
		}
	}

	return &apiService.BackupConfigCreateCreated{
		Payload: util.BackupConfigResourceToModel(secretResource),
	}
}
