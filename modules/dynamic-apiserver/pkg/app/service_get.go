package app

import (
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
)

func (h *handlers) ServiceGetHandler(params apiService.ServiceGetParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	result, err := h.Services().Get(ctx, params.ServiceID, v1.GetOptions{})
	if errors.IsNotFound(err) {
		msg := fmt.Sprintf("kuberlogic service not found: %s", params.ServiceID)
		h.log.Warnw(msg, "error", err)
		return apiService.NewServiceGetNotFound().WithPayload(&models.Error{
			Message: msg,
		})
	} else if err != nil {
		msg := "error finding service"
		h.log.Errorw(msg, "error", err)
		return apiService.NewServiceGetServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}

	service, err := util.KuberlogicToService(result)
	if err != nil {
		h.log.Errorw("error converting kuberlogicservice", "error", err)
		return apiService.NewServiceGetServiceUnavailable().WithPayload(
			&models.Error{
				Message: err.Error(),
			})
	}

	return apiService.NewServiceGetOK().WithPayload(service)
}
