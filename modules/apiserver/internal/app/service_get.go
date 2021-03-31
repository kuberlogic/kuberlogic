package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/internal/security"
	"github.com/kuberlogic/operator/modules/apiserver/util"
)

func (srv *Service) ServiceGetHandler(params apiService.ServiceGetParams, principal *models.Principal) middleware.Responder {
	ns, name, err := util.SplitID(params.ServiceID)
	if err != nil {
		util.BadRequestFromError(err)
	}

	if authorized, err := srv.authProvider.Authorize(principal.Token, security.ServiceGetSecGrant, params.ServiceID); err != nil {
		srv.log.Errorw("error checking authorization", "error", err)
		resp := apiService.NewServiceGetServiceUnavailable().WithPayload(&models.Error{Message: "eror checking authorization"})
		return resp
	} else if !authorized {
		resp := apiService.NewServiceGetForbidden()
		return resp
	}

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
