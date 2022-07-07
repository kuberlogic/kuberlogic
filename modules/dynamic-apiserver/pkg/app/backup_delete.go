package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiBackup "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/backup"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (srv *Service) BackupDeleteHandler(params apiBackup.BackupDeleteParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	if err := srv.kuberlogicClient.Delete().
		Resource(backupK8sResource).
		Name(params.BackupID).
		Do(ctx).
		Error(); errors.IsNotFound(err) {
		return apiBackup.NewBackupDeleteNotFound()
	} else if err != nil {
		srv.log.Errorw("error deleting klb", "error", err, "name", params.BackupID)
		return apiBackup.NewBackupDeleteServiceUnavailable().WithPayload(&models.Error{
			Message: "error deleting backup",
		})
	}
	return apiBackup.NewBackupDeleteOK()
}
