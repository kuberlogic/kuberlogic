package app

import (
	"github.com/go-openapi/runtime/middleware"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiBackup "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/backup"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
)

func (h *handlers) BackupListHandler(params apiBackup.BackupListParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	opts := h.ListOptionsByKeyValue(util.BackupRestoreServiceField, params.ServiceID)
	r, err := h.Backups().List(ctx, opts)
	if err != nil {
		msg := "error listing backups"
		h.log.Errorw(msg)
		return apiBackup.NewBackupListServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}
	h.log.Debugw("found kuberlogicservicebackups objects", "count", len(r.Items), "objects", r)

	items := make([]*models.Backup, 0)
	for _, klb := range r.Items {
		b, err := util.KuberlogicToBackup(&klb)
		if err != nil {
			h.log.Errorw("error converting klb to model", "error", err, "name", klb.GetName())
			return apiBackup.NewBackupListServiceUnavailable().WithPayload(&models.Error{
				Message: "error converting backup object to model",
			})
		}
		items = append(items, b)
	}
	return apiBackup.NewBackupListOK().WithPayload(items)
}
