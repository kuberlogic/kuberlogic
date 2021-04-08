package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
)

func (srv *Service) ServiceDeleteHandler(params apiService.ServiceDeleteParams, principal *models.Principal) middleware.Responder {
	ns, name := srv.existingService.Namespace, srv.existingService.Name

	s := srv.serviceStore.NewServiceObject(name, ns)
	errDelete := srv.serviceStore.DeleteService(s, params.HTTPRequest.Context())
	if errDelete != nil {
		srv.log.Errorw("service delete error", "namespace", ns, "name", name, "error", errDelete.Err)
		return apiService.NewServiceDeleteServiceUnavailable().WithPayload(&models.Error{Message: errDelete.ClientMsg})
	}

	return &apiService.ServiceDeleteOK{}
}
