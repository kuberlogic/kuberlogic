package app

import (
	"github.com/go-openapi/runtime/middleware"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/restapi/operations/service"
)

// set this string to a required security grant for this action
const serviceAddSecGrant = "nonsense"

func (srv *Service) ServiceAddHandler(params apiService.ServiceAddParams) middleware.Responder {

	return middleware.NotImplemented("operation service ServiceAdd has not yet been implemented")
}
