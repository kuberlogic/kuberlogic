package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
)

func (srv *Service) LogsGetHandler(params apiService.LogsGetParams, principal *models.Principal) middleware.Responder {
	ns, name := srv.existingService.Namespace, srv.existingService.Name

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
