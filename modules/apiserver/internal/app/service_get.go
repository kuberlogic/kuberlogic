package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
)

func (srv *Service) ServiceGetHandler(params apiService.ServiceGetParams, principal *models.Principal) middleware.Responder {
	kuberlogicService := params.HTTPRequest.Context().Value("service").(*kuberlogicv1.KuberLogicService)
	ns, name := kuberlogicService.Namespace, kuberlogicService.Name
	// TODO: need to use kuberLogicToService directly

	service, found, errGet := srv.serviceStore.GetService(name, ns, params.HTTPRequest.Context())
	if errGet != nil {
		srv.log.Errorw("service get error", "error", errGet.Err)
		return apiService.NewServiceGetServiceUnavailable().WithPayload(&models.Error{Message: errGet.ClientMsg})
	}
	if !found {
		return apiService.NewServiceGetNotFound()
	}

	return apiService.NewServiceGetOK().WithPayload(service)
}
