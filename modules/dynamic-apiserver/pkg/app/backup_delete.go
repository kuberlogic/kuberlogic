package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiBackup "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/backup"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h *handlers) BackupDeleteHandler(params apiBackup.BackupDeleteParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	if err := h.Backups().Delete(ctx, params.BackupID, v1.DeleteOptions{}); errors.IsNotFound(err) {
		return apiBackup.NewBackupDeleteNotFound().WithPayload(&models.Error{
			Message: "backup not found: " + params.BackupID,
		})
	} else if err != nil {
		h.log.Errorw("error deleting klb", "error", err, "name", params.BackupID)
		return apiBackup.NewBackupDeleteServiceUnavailable().WithPayload(&models.Error{
			Message: "error deleting backup",
		})
	}
	return apiBackup.NewBackupDeleteOK()
}
