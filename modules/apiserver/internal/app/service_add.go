package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
)

func (srv *Service) ServiceAddHandler(params apiService.ServiceAddParams, principal *models.Principal) middleware.Responder {
	svc, errCreate := srv.serviceStore.CreateService(params.ServiceItem, principal.Email, params.HTTPRequest.Context())
	if errCreate != nil {
		srv.log.Errorw("service create error", "error", errCreate.Err)
		if errCreate.Client {
			return apiService.NewServiceAddBadRequest().WithPayload(&models.Error{Message: errCreate.ClientMsg})
		} else {
			return apiService.NewServiceAddServiceUnavailable().WithPayload(&models.Error{Message: errCreate.ClientMsg})
		}
	}

	return apiService.NewServiceAddCreated().WithPayload(svc)
}
