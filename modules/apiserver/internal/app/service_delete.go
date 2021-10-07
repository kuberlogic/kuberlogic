package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/restapi/operations/service"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
)

func (srv *Service) ServiceDeleteHandler(params apiService.ServiceDeleteParams, principal *models.Principal) middleware.Responder {
	service := params.HTTPRequest.Context().Value("service").(*kuberlogicv1.KuberLogicService)
	ns, name := principal.Namespace, service.Name

	s := srv.serviceStore.NewServiceObject(name, ns)
	errDelete := srv.serviceStore.DeleteService(s, principal, params.HTTPRequest.Context())
	if errDelete != nil {
		srv.log.Errorw("service delete error", "namespace", ns, "name", name, "error", errDelete.Err)
		return apiService.NewServiceDeleteServiceUnavailable().WithPayload(&models.Error{Message: errDelete.ClientMsg})
	}

	return &apiService.ServiceDeleteOK{}
}
