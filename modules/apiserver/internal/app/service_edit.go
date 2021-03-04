package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
)

// set this string to a required security grant for this action
const serviceEditSecGrant = "service:edit"

func (srv *Service) ServiceEditHandler(params apiService.ServiceEditParams, principal *models.Principal) middleware.Responder {
	if authorized, err := srv.authProvider.Authorize(principal.Token, serviceEditSecGrant, params.ServiceID); err != nil {
		srv.log.Errorf("error checking authorization " + err.Error())
		resp := apiService.NewServiceEditServiceUnavailable().WithPayload(&models.Error{Message: "error checking authorization"})
		return resp
	} else if !authorized {
		resp := apiService.NewServiceEditForbidden()
		return resp
	}

	m, errUpdate := srv.serviceStore.UpdateService(params.ServiceItem, params.HTTPRequest.Context())
	if errUpdate != nil {
		srv.log.Errorf("error updating service: %s", errUpdate.Err.Error())
		if errUpdate.Client {
			return apiService.NewServiceEditBadRequest().WithPayload(&models.Error{Message: errUpdate.ClientMsg})
		} else {
			return apiService.NewServiceEditServiceUnavailable().WithPayload(&models.Error{Message: errUpdate.ClientMsg})
		}
	}
	return apiService.NewServiceEditOK().WithPayload(m)
}
