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
		srv.log.Errorf("error checking authorization " + err.Error())
		resp := apiService.NewBackupConfigDeleteBadRequest()
		return resp
	} else if !authorized {
		resp := apiService.NewBackupConfigDeleteForbidden()
		return resp
	}

	srv.log.Debugf("attempting to delete a backup config %s/%s", ns, name)
	err = srv.clientset.CoreV1().Secrets(ns).
		Delete(context.TODO(), name, v1.DeleteOptions{})
	if errors.IsNotFound(err) {
		srv.log.Errorf("backup config %s/%s does not exist: %s", ns, name, err.Error())
		return &apiService.BackupConfigDeleteNotFound{}
	}
	if err != nil {
		srv.log.Errorf("error deleting backup config: %s", err.Error())
		return util.BadRequestFromError(err)
	}

	srv.log.Debugf("attempting to delete a backup resource %s/%s", ns, name)
	err = util.DeleteBackupResource(srv.cmClient, ns, name)
	if err != nil {
		srv.log.Errorf("error deleting backup resource: %s", err.Error())
		return util.BadRequestFromError(err)
	}

	return &apiService.BackupConfigDeleteOK{}
}
