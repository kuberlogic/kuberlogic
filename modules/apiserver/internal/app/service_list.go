package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
)

func (srv *Service) ServiceListHandler(params apiService.ServiceListParams, principal *models.Principal) middleware.Responder {
	services, errList := srv.serviceStore.ListServices(params.HTTPRequest.Context())
	if errList != nil {
		srv.log.Errorw("list services error", "error", errList.Err)
		return apiService.NewServiceListServiceUnavailable().WithPayload(&models.Error{Message: errList.ClientMsg})
	}

	return apiService.NewServiceListOK().WithPayload(services)
}
