package app

import (
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiRestore "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/restore"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
)

func (h *handlers) RestoreAddHandler(params apiRestore.RestoreAddParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	backupName := params.RestoreItem.BackupID
	klb, err := h.Backups().Get(ctx, backupName, v1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return apiRestore.NewRestoreAddBadRequest().WithPayload(&models.Error{
			Message: fmt.Sprintf("backup `%s` not found", backupName),
		})
	} else if err != nil {
		h.log.Errorw("error getting kuberlogicservicebackup for restore", "error", err)
		return apiRestore.NewRestoreAddServiceUnavailable().WithPayload(&models.Error{
			Message: fmt.Sprintf("error getting coresponding backup %s: %s", backupName, err),
		})
	}

	klr := util.RestoreToKuberlogic(params.RestoreItem, klb)
	klr.SetName(klb.GetName())

	result, err := h.Restores().Create(ctx, klr, v1.CreateOptions{})
	if k8serrors.IsAlreadyExists(err) {
		h.log.Errorw("klr already exists", "name", klr.GetName())
		return apiRestore.NewRestoreAddConflict()
	} else if err != nil {
		h.log.Errorw("error creating klr", "error", err, "name", klr.GetName())
		return apiRestore.NewRestoreAddServiceUnavailable().WithPayload(&models.Error{
			Message: err.Error(),
		})
	}
	return apiRestore.NewRestoreAddCreated().WithPayload(util.KuberlogicToRestore(result))
}
