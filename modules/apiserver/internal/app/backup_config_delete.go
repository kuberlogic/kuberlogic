package app

import (
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/util"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// set this string to a required security grant for this action
const backupConfigDeleteSecGrant = "service:backup-config:delete"

func (srv *Service) BackupConfigDeleteHandler(params apiService.BackupConfigDeleteParams, principal *models.Principal) middleware.Responder {
	// validate path parameter
	ns, name, err := util.SplitID(params.ServiceID)
	if err != nil {
		return util.BadRequestFromError(err)
	}

	if authorized, err := srv.authProvider.Authorize(principal.Token, backupConfigDeleteSecGrant, params.ServiceID); err != nil {
		srv.log.Errorw("error checking authorization", "error", err)
		resp := apiService.NewBackupConfigDeleteBadRequest()
		return resp
	} else if !authorized {
		resp := apiService.NewBackupConfigDeleteForbidden()
		return resp
	}

	srv.log.Debugw("attempting to delete a backup config",
		"namespace", ns, "name", name)
	err = srv.clientset.CoreV1().Secrets(ns).
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
	err = util.DeleteBackupResource(srv.cmClient, ns, name)
	if err != nil {
		srv.log.Errorw("error deleting backup resource", "error", err)
		return util.BadRequestFromError(err)
	}

	return &apiService.BackupConfigDeleteOK{}
}
