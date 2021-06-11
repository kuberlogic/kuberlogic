package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
)

func (srv *Service) LogsGetHandler(params apiService.LogsGetParams, principal *models.Principal) middleware.Responder {
	service := params.HTTPRequest.Context().Value("service").(*kuberlogicv1.KuberLogicService)
	ns, name := principal.Namespace, service.Name

	m := srv.serviceStore.NewServiceObject(name, ns)
	logs, errLogs := srv.serviceStore.GetServiceLogs(m, params.ServiceInstance, *params.Tail, params.HTTPRequest.Context())
	if errLogs != nil {
		srv.log.Errorw("error getting service logs", "error", errLogs.Err)
		return apiService.NewLogsGetServiceUnavailable().WithPayload(&models.Error{Message: errLogs.ClientMsg})
	}

	return apiService.NewLogsGetOK().WithPayload(&models.Log{
		Body:  logs,
		Lines: *params.Tail,
	})
}
