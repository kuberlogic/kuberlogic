package app

import (
	"github.com/go-openapi/runtime/middleware"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/restapi/operations/service"
)

// set this string to a required security grant for this action
const serviceGetSecGrant = "nonsense"

func (srv *Service) ServiceGetHandler(params apiService.ServiceGetParams) middleware.Responder {

	return middleware.NotImplemented("operation service ServiceGet has not yet been implemented")
}
