package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/util"
)

// set this string to a required security grant for this action
const logsGetSecGrant = "service:logs"

func (srv *Service) LogsGetHandler(params apiService.LogsGetParams, principal *models.Principal) middleware.Responder {
	ns, name, err := util.SplitID(params.ServiceID)
	if err != nil {
		return util.BadRequestFromError(err)
	}

	if authorized, err := srv.authProvider.Authorize(principal.Token, logsGetSecGrant, params.ServiceID); err != nil {
		srv.log.Errorf("error checking authorization " + err.Error())
		return apiService.NewLogsGetForbidden()
	} else if !authorized {
		return apiService.NewLogsGetForbidden()
	}

	m := srv.serviceStore.NewServiceObject(name, ns)
	logs, errLogs := srv.serviceStore.GetServiceLogs(m, params.ServiceInstance, *params.Tail, params.HTTPRequest.Context())
	if errLogs != nil {
		srv.log.Errorf("error getting service logs: %s", errLogs.Err.Error())
		return apiService.NewLogsGetServiceUnavailable().WithPayload(&models.Error{Message: errLogs.ClientMsg})
	}

	return apiService.NewLogsGetOK().WithPayload(&models.Log{
		Body:  logs,
		Lines: *params.Tail,
	})
}
