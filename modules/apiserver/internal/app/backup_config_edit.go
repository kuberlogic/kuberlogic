package app

import (
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/util"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// set this string to a required security grant for this action
const backupConfigEditSecGrant = "service:backup-config:edit"

// curl -v -H Content-Type:application/json -H "Authorization: Bearer" -X PUT localhost:8001/api/v1/services/<service-id>/backup-config -d '{"aws_access_key_id":"","aws_secret_access_key":"","bucket":"cloudmanaged","endpoint":"https://fra1.digitaloceanspaces.com","schedule":"* 1 * * *","type":"s3","enabled":false}'
func (srv *Service) BackupConfigEditHandler(params apiService.BackupConfigEditParams, principal *models.Principal) middleware.Responder {
	// validation
	ns, name, err := util.SplitID(params.ServiceID)
	if err != nil {
		return util.BadRequestFromError(err)
	}

	if authorized, err := srv.authProvider.Authorize(principal.Token, backupConfigEditSecGrant, params.ServiceID); err != nil {
		srv.log.Errorw("error checking authorization", "error", err)
		resp := apiService.NewBackupConfigEditBadRequest()
		return resp
	} else if !authorized {
		resp := apiService.NewBackupConfigEditForbidden()
		return resp
	}

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
		err = util.CreateBackupResource(srv.cmClient, ns, name, *params.BackupConfig.Schedule)
		if err != nil {
			srv.log.Errorw("error create a backup resource", "error", err)
			return util.BadRequestFromError(err)
		}

		srv.log.Debugw("attempting to update a backup resource",
			"namespace", ns, "name", name)
		err = util.UpdateBackupResource(srv.cmClient, ns, name, *params.BackupConfig.Schedule)
		if err != nil {
			srv.log.Errorw("error update a backup resource", "error", err)
			return util.BadRequestFromError(err)
		}
	} else {
		srv.log.Debugw("attempting to delete a backup resource",
			"namespace", ns, "name", name)
		err = util.DeleteBackupResource(srv.cmClient, ns, name)
		if err != nil {
			srv.log.Errorw("error deleting backup resource", "error", err)
			return util.BadRequestFromError(err)
		}
	}

	return &apiService.BackupConfigEditOK{
		Payload: util.BackupConfigResourceToModel(updatedResource),
	}
}
