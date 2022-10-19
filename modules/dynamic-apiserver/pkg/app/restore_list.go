package app

import (
	"github.com/go-openapi/runtime/middleware"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiRestore "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/restore"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
)

func (h *handlers) RestoreListHandler(params apiRestore.RestoreListParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	opts := h.ListOptionsByKeyValue(util.BackupRestoreServiceField, params.ServiceID)
	result, err := h.Restores().List(ctx, opts)
	if err != nil {
		msg := "error listing result"
		h.log.Errorw(msg)
		return apiRestore.NewRestoreListServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}
	h.log.Debugw("found kuberlogicservicerestores objects", "count", len(result.Items), "objects", result)

	items := make([]*models.Restore, 0)
	for _, klr := range result.Items {
		items = append(items, util.KuberlogicToRestore(&klr))
	}
	return apiRestore.NewRestoreListOK().WithPayload(items)
}
