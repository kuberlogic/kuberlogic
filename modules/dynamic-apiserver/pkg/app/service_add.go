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

func (h *handlers) ServiceAddHandler(params apiService.ServiceAddParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	if params.ServiceItem.Subscription != "" {
		opts := h.ListOptionsByKeyValue(util.SubscriptionField, &params.ServiceItem.Subscription)
		if found, err := h.Services().Exists(ctx, opts); err != nil {
			return apiService.NewServiceAddServiceUnavailable().WithPayload(
				&models.Error{
					Message: err.Error(),
				})
		} else if found {
			return apiService.NewServiceAddBadRequest().WithPayload(
				&models.Error{
					Message: fmt.Sprintf("service with subscription '%s' already exist", params.ServiceItem.Subscription),
				})
		}
	}

	c, err := util.ServiceToKuberlogic(params.ServiceItem, h.config)
	if err != nil {
		h.log.Errorw("error converting service model to kuberlogic", "error", err)
		return apiService.NewServiceAddBadRequest().WithPayload(
			&models.Error{
				Message: err.Error(),
			})
	}

	result, err := h.Services().Create(ctx, c, v1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		msg := fmt.Sprintf("kuberlogic service already exists: %s", *params.ServiceItem.ID)
		h.log.Warnw(msg, "error", err)
		return apiService.NewServiceAddConflict()
	} else if err != nil {
		h.log.Errorw("error creating kuberlogicservice", "error", err)
		return apiService.NewServiceAddServiceUnavailable().WithPayload(
			&models.Error{
				Message: err.Error(),
			})
	}
	svc, err := util.KuberlogicToService(result)
	if err != nil {
		h.log.Errorw("error converting kuberlogicservice to model", "error", err)
		return apiService.NewServiceAddBadRequest().WithPayload(
			&models.Error{
				Message: err.Error(),
			})
	}

	return apiService.NewServiceAddCreated().WithPayload(svc)
}
