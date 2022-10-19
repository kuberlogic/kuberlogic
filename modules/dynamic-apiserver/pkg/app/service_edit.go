package app

import (
	"encoding/json"
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
)

func (h *handlers) ServiceEditHandler(params apiService.ServiceEditParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	// TODO: check if service exists by params.ServiceID

	if params.ServiceItem.Subscription != "" {
		return apiService.NewServiceEditBadRequest().WithPayload(
			&models.Error{
				Message: "subscription cannot be changed",
			})
	}

	c, err := util.ServiceToKuberlogic(params.ServiceItem, h.config)
	if err != nil {
		h.log.Errorw("error converting service model to kuberlogic", "error", err)
		return apiService.NewServiceEditBadRequest().WithPayload(
			&models.Error{
				Message: err.Error(),
			})
	}

	patch, err := json.Marshal(c)
	if err != nil {
		h.log.Errorw("service decode error", "error", err)
		return apiService.NewServiceEditBadRequest().WithPayload(
			&models.Error{
				Message: err.Error(),
			})
	}

	if _, err = h.Services().Patch(ctx, c.GetName(), types.MergePatchType, patch, v1.PatchOptions{}); errors.IsNotFound(err) {
		msg := fmt.Sprintf("kuberlogic service not found: %s", params.ServiceID)
		h.log.Warnw(msg, "error", err)
		return apiService.NewServiceEditNotFound().WithPayload(&models.Error{
			Message: msg,
		})
	} else if err != nil {
		h.log.Errorw("error creating kuberlogicservice", "error", err)
		return apiService.NewServiceEditBadRequest().WithPayload(
			&models.Error{
				Message: err.Error(),
			})
	}

	return apiService.NewServiceEditOK()
}
