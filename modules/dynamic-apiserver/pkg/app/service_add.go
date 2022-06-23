package app

import (
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (srv *Service) ServiceAddHandler(params apiService.ServiceAddParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	c, err := util.ServiceToKuberlogic(params.ServiceItem)
	if err != nil {
		srv.log.Errorw("error converting service model to kuberlogic", "error", err)
		return apiService.NewServiceAddBadRequest().WithPayload(
			&models.Error{
				Message: err.Error(),
			})
	}

	result := new(kuberlogiccomv1alpha1.KuberLogicService)
	err = srv.kuberlogicClient.Post().
		Resource(serviceK8sResource).
		Name(*params.ServiceItem.ID).
		Body(c).
		Do(ctx).
		Into(result)
	if err != nil && util.CheckStatus(err, v1.StatusReasonAlreadyExists) {
		msg := fmt.Sprintf("kuberlogic service already exists: %s", *params.ServiceItem.ID)
		srv.log.Warnw(msg, "error", err)
		return apiService.NewServiceAddConflict()
	} else if err != nil {
		srv.log.Errorw("error creating kuberlogicservice", "error", err)
		return apiService.NewServiceAddBadRequest().WithPayload(
			&models.Error{
				Message: err.Error(),
			})
	}
	svc, err := util.KuberlogicToService(result)
	if err != nil {
		srv.log.Errorw("error converting kuberlogicservice to model", "error", err)
		return apiService.NewServiceAddBadRequest().WithPayload(
			&models.Error{
				Message: err.Error(),
			})
	}

	return apiService.NewServiceAddCreated().WithPayload(svc)
}
