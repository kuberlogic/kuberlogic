package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
)

func (srv *Service) ServiceListHandler(params apiService.ServiceListParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	res, err := srv.ListKuberlogicServicesBySubscription(ctx, params.SubscriptionID)
	if err != nil {
		msg := "error listing service"
		srv.log.Errorw(msg)
		return apiService.NewServiceListServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}
	srv.log.Debugw("found kuberlogicservice objects", "length", len(res.Items), "objects", res)

	var services []*models.Service
	for _, r := range res.Items {
		service, err := util.KuberlogicToService(&r)
		if err != nil {
			msg := "error converting service object"
			srv.log.Errorw(msg)
			return apiService.NewServiceListServiceUnavailable().WithPayload(&models.Error{
				Message: msg,
			})
		}
		services = append(services, service)
	}

	return apiService.NewServiceListOK().WithPayload(services)
}
