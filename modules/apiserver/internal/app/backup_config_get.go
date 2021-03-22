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
const backupConfigGetSecGrant = "service:backup-config:get"

// curl -v -H Content-Type:application/json -H "Authorization: Bearer" -X GET localhost:8001/api/v1/services/<service-id>/backup-config
func (srv *Service) BackupConfigGetHandler(params apiService.BackupConfigGetParams, principal *models.Principal) middleware.Responder {
	ns, name, err := util.SplitID(params.ServiceID)
	if err != nil {
		return util.BadRequestFromError(err)
	}

	if authorized, err := srv.authProvider.Authorize(principal.Token, backupConfigGetSecGrant, params.ServiceID); err != nil {
		srv.log.Errorf("error checking authorization: %s ", err.Error())
		resp := apiService.NewBackupConfigEditBadRequest()
		return resp
	} else if !authorized {
		resp := apiService.NewBackupConfigEditForbidden()
		return resp
	}

	srv.log.Debugf("attempting to get a backup config %s/%s", ns, name)
	secret, err := srv.clientset.CoreV1().
		Secrets(ns).
		Get(context.TODO(), name, v1.GetOptions{})
	if errors.IsNotFound(err) {
		srv.log.Errorf("backup config %s/%s does not exist: %s", ns, name, err.Error())
		return &apiService.BackupConfigGetNotFound{}
	} else if err != nil {
		srv.log.Errorf("failed to get a backup config %s/%s: %s", ns, name, err.Error())
		return util.BadRequestFromError(err)
	}

	//srv.log.Infof("aws-secret-key: %s", secret.Data["aws-access-key-id"])
	//srv.log.Infof("bucket: %s", secret.Data["bucket"])
	return &apiService.BackupConfigGetOK{
		Payload: util.BackupConfigResourceToModel(secret),
	}
}
