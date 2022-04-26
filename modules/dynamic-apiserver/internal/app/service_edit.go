package app

import (
	"github.com/go-openapi/runtime/middleware"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/restapi/operations/service"
)

// set this string to a required security grant for this action
const serviceEditSecGrant = "nonsense"

func (srv *Service) ServiceEditHandler(params apiService.ServiceEditParams) middleware.Responder {

	return middleware.NotImplemented("operation service ServiceEdit has not yet been implemented")
}
