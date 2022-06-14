package app

import (
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/util"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func (srv *Service) ServiceGetHandler(params apiService.ServiceGetParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	result := new(kuberlogiccomv1alpha1.KuberLogicService)
	err := srv.kuberlogicClient.Get().
		Resource(serviceK8sResource).
		Name(params.ServiceID).
		Do(ctx).
		Into(result)
	if err != nil && util.ErrNotFound(err) {
		msg := fmt.Sprintf("kuberlogic service not found: %s", params.ServiceID)
		srv.log.Warnw(msg, "error", err)
		return apiService.NewServiceGetNotFound().WithPayload(&models.Error{
			Message: msg,
		})
	} else if err != nil {
		msg := "error finding service"
		srv.log.Errorw(msg, "error", err)
		return apiService.NewServiceGetServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}

	service, err := util.KuberlogicToService(result)
	if err != nil {
		srv.log.Errorw("error converting kuberlogicservice", "error", err)
		return apiService.NewServiceGetServiceUnavailable().WithPayload(
			&models.Error{
				Message: err.Error(),
			})
	}

	return apiService.NewServiceGetOK().WithPayload(service)
}
