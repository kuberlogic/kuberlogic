package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/util"
)

// set this string to a required security grant for this action
const serviceGetSecGrant = "service:get"

func (srv *Service) ServiceGetHandler(params apiService.ServiceGetParams, principal *models.Principal) middleware.Responder {
	ns, name, err := util.SplitID(params.ServiceID)
	if err != nil {
		util.BadRequestFromError(err)
	}

	if authorized, err := srv.authProvider.Authorize(principal.Token, serviceGetSecGrant, params.ServiceID); err != nil {
		srv.log.Errorf("error checking authorization " + err.Error())
		resp := apiService.NewServiceGetServiceUnavailable().WithPayload(&models.Error{Message: "eror checking authorization"})
		return resp
	} else if !authorized {
		resp := apiService.NewServiceGetForbidden()
		return resp
	}

	service, found, errGet := srv.serviceStore.GetService(name, ns, params.HTTPRequest.Context())
	if errGet != nil {
		srv.log.Errorf("service get error: %s", errGet.Err.Error())
		return apiService.NewServiceGetServiceUnavailable().WithPayload(&models.Error{Message: errGet.ClientMsg})
	}
	if !found {
		return apiService.NewServiceGetNotFound()
	}

	return apiService.NewServiceGetOK().WithPayload(service)
}
