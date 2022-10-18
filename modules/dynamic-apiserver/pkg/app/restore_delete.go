package app

import (
	"github.com/go-openapi/runtime/middleware"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiRestore "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/restore"
)

func (h *handlers) RestoreDeleteHandler(params apiRestore.RestoreDeleteParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	if err := h.Restores().Delete(ctx, params.RestoreID, v1.DeleteOptions{}); errors.IsNotFound(err) {
		return apiRestore.NewRestoreDeleteNotFound()
	} else if err != nil {
		h.log.Errorw("error deleting klr", "error", err, "name", params.RestoreID)
		return apiRestore.NewRestoreDeleteServiceUnavailable().WithPayload(&models.Error{
			Message: "error deleting restore",
		})
	}
	return apiRestore.NewRestoreDeleteOK()
}
