package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/util"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

// set this string to a required security grant for this action
const serviceListSecGrant = "nonsense"

func (srv *Service) ServiceListHandler(params apiService.ServiceListParams) middleware.Responder {
	ctx := params.HTTPRequest.Context()
	res := new(kuberlogiccomv1alpha1.KuberLogicServiceList)

	err := srv.kuberlogicClient.Get().
		Resource(serviceK8sResource).
		//Namespace(p.Namespace). --- ?
		Do(ctx).
		Into(res)
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
