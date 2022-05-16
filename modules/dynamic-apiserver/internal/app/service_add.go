package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/util"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

// set this string to a required security grant for this action
const serviceAddSecGrant = "nonsense"

func (srv *Service) ServiceAddHandler(params apiService.ServiceAddParams) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	c, err := util.ServiceToKuberlogic(params.ServiceItem)
	if err != nil {
		return apiService.NewServiceAddBadRequest().WithPayload(
			&models.Error{
				Message: err.Error(),
			})
	}

	result := new(kuberlogiccomv1alpha1.KuberLogicService)
	err = srv.kuberlogicClient.Post().
		Resource(serviceK8sResource).
		Namespace(params.ServiceItem.Ns).
		Name(*params.ServiceItem.Name).
		Body(c).
		Do(ctx).
		Into(result)
	if err != nil {
		return apiService.NewServiceAddBadRequest().WithPayload(
			&models.Error{
				Message: err.Error(),
			})
	}
	svc, err := util.KuberlogicToService(result)
	if err != nil {
		return apiService.NewServiceAddBadRequest().WithPayload(
			&models.Error{
				Message: err.Error(),
			})
	}

	return apiService.NewServiceAddCreated().WithPayload(svc)
}
