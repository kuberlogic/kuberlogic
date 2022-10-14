package app

import (
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
)

func (h *handlers) ServiceDeleteHandler(params apiService.ServiceDeleteParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	if _, err := h.Services().Get(ctx, params.ServiceID, v1.GetOptions{}); errors.IsNotFound(err) {
		msg := fmt.Sprintf("kuberlogic service not found: %s", params.ServiceID)
		h.log.Warnw(msg, "error", err)
		return apiService.NewServiceDeleteNotFound().WithPayload(&models.Error{
			Message: msg,
		})
	} else if err != nil {
		msg := "error finding service"
		h.log.Errorw(msg, "error", err)
		return apiService.NewServiceDeleteServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}

	if err := h.Services().Delete(ctx, params.ServiceID, v1.DeleteOptions{}); err != nil {
		msg := "error deleting service"
		h.log.Errorw(msg, "error", err)
		return apiService.NewServiceDeleteServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}

	return apiService.NewServiceDeleteOK()
}
