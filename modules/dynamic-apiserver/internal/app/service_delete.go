package app

import (
	"github.com/go-openapi/runtime/middleware"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/restapi/operations/service"
)

// set this string to a required security grant for this action
const serviceDeleteSecGrant = "nonsense"

func (srv *Service) ServiceDeleteHandler(params apiService.ServiceDeleteParams) middleware.Responder {

	return middleware.NotImplemented("operation service ServiceDelete has not yet been implemented")
}
