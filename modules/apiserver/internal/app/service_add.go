package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/util"
)

const (
	serviceAddSecGrant = "service:add"
)

func (srv *Service) ServiceAddHandler(params apiService.ServiceAddParams, principal *models.Principal) middleware.Responder {
	name := params.ServiceItem.Name
	ns := params.ServiceItem.Ns
	id, err := util.JoinID(*ns, *name)
	if err != nil {
		return util.BadRequestFromError(err)
	}

	if authorized, err := srv.authProvider.Authorize(principal.Token, serviceAddSecGrant, id); err != nil {
		srv.log.Errorw("error checking authorization", "error", err)
		resp := apiService.NewServiceAddServiceUnavailable().WithPayload(&models.Error{Message: "error checking authorization"})
		return resp
	} else if !authorized {
		resp := apiService.NewServiceAddForbidden()
		return resp
	}

	svc, errCreate := srv.serviceStore.CreateService(params.ServiceItem, params.HTTPRequest.Context())
	if errCreate != nil {
		srv.log.Errorw("service create error", errCreate.Err)
		if errCreate.Client {
			return apiService.NewServiceAddBadRequest().WithPayload(&models.Error{Message: errCreate.ClientMsg})
		} else {
			return apiService.NewServiceAddServiceUnavailable().WithPayload(&models.Error{Message: errCreate.ClientMsg})
		}
	}

	return apiService.NewServiceAddCreated().WithPayload(svc)
}
