package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/internal/security"
	"github.com/kuberlogic/operator/modules/apiserver/util"
)

func (srv *Service) ServiceDeleteHandler(params apiService.ServiceDeleteParams, principal *models.Principal) middleware.Responder {
	ns, name, err := util.SplitID(params.ServiceID)
	if err != nil {
		return util.BadRequestFromError(err)
	}

	if authorized, err := srv.authProvider.Authorize(principal.Token, security.ServiceDeleteSecGrant, params.ServiceID); err != nil {
		srv.log.Errorw("error checking authorization", "error", err)
		resp := apiService.NewServiceDeleteBadRequest().WithPayload(&models.Error{Message: "error checking authorization"})
		return resp
	} else if !authorized {
		resp := apiService.NewServiceDeleteForbidden()
		return resp
	}

	// clean permissions first
	if err := srv.authProvider.DeletePermissionResource(params.ServiceID); err != nil {
		srv.log.Errorw("error deleting associated permission resource", "error", err, "service", params.ServiceID)
		return apiService.NewServiceDeleteServiceUnavailable().WithPayload(&models.Error{Message: "error deleting associated permission resource"})
	}
	s := srv.serviceStore.NewServiceObject(name, ns)
	errDelete := srv.serviceStore.DeleteService(s, params.HTTPRequest.Context())
	if errDelete != nil {
		srv.log.Errorw("service delete error", "namespace", ns, "name", name, "error", err)
		return apiService.NewServiceDeleteServiceUnavailable().WithPayload(&models.Error{Message: errDelete.ClientMsg})
	}

	return &apiService.ServiceDeleteOK{}
}
