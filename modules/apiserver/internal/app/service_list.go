package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
)

// set this string to a required security grant for this action
const serviceListSecGrant = "service:list"

func (srv *Service) ServiceListHandler(params apiService.ServiceListParams, principal *models.Principal) middleware.Responder {

	if authorized, err := srv.authProvider.Authorize(principal.Token, serviceListSecGrant, "*"); err != nil {
		srv.log.Errorf("error checking authorization " + err.Error())
		resp := apiService.NewServiceListServiceUnavailable().WithPayload(&models.Error{Message: "error checking authorization"})
		return resp
	} else if !authorized {
		resp := apiService.NewServiceListForbidden()
		return resp
	}

	services, errList := srv.serviceStore.ListServices(params.HTTPRequest.Context())
	if errList != nil {
		srv.log.Errorf("list services error: %s", errList.Err.Error())
		return apiService.NewServiceListServiceUnavailable().WithPayload(&models.Error{Message: errList.ClientMsg})
	}

	return apiService.NewServiceListOK().WithPayload(services)
}
