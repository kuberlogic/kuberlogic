package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiRestore "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/restore"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
)

func (h *handlers) RestoreListHandler(params apiRestore.RestoreListParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	klrs, err := h.Restores().ListByServiceName(ctx, params.ServiceID)
	if err != nil {
		msg := "error listing restores"
		h.log.Errorw(msg)
		return apiRestore.NewRestoreListServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}
	h.log.Debugw("found kuberlogicservicerestores objects", "count", len(klrs.Items), "objects", klrs)

	items := make([]*models.Restore, 0)
	for _, klr := range klrs.Items {
		b, err := util.KuberlogicToRestore(&klr)
		if err != nil {
			h.log.Errorw("error converting klr to model", "error", err, "name", klr.GetName())
			return apiRestore.NewRestoreListServiceUnavailable().WithPayload(&models.Error{
				Message: "error converting restore object to model",
			})
		}
		items = append(items, b)
	}
	return apiRestore.NewRestoreListOK().WithPayload(items)
}
