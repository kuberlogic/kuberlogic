package app

import (
	"github.com/go-openapi/runtime/middleware"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/restapi/operations/service"
)

// set this string to a required security grant for this action
const serviceListSecGrant = "nonsense"

func (srv *Service) ServiceListHandler(params apiService.ServiceListParams) middleware.Responder {

	return middleware.NotImplemented("operation service ServiceList has not yet been implemented")
}
