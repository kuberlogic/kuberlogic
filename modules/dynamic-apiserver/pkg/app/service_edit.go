package app

import (
	"encoding/json"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (srv *Service) ServiceEditHandler(params apiService.ServiceEditParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	if params.ServiceItem.Subscription != "" {
		return apiService.NewServiceEditBadRequest().WithPayload(
			&models.Error{
				Message: fmt.Sprintf("subscription cannot be changed"),
			})
	}

	c, err := util.ServiceToKuberlogic(params.ServiceItem)
	if err != nil {
		srv.log.Errorw("error converting service model to kuberlogic", "error", err)
		return apiService.NewServiceEditBadRequest().WithPayload(
			&models.Error{
				Message: err.Error(),
			})
	}

	patch, err := json.Marshal(c)
	if err != nil {
		srv.log.Errorw("service decode error", "error", err)
		return apiService.NewServiceEditBadRequest().WithPayload(
			&models.Error{
				Message: err.Error(),
			})
	}

	result := new(kuberlogiccomv1alpha1.KuberLogicService)
	err = srv.kuberlogicClient.Patch(types.MergePatchType).
		Resource(serviceK8sResource).
		Name(*params.ServiceItem.ID).
		Body(patch).
		Do(ctx).
		Into(result)
	if err != nil && util.CheckStatus(err, v1.StatusReasonNotFound) {
		msg := fmt.Sprintf("kuberlogic service not found: %s", params.ServiceID)
		srv.log.Warnw(msg, "error", err)
		return apiService.NewServiceEditNotFound().WithPayload(&models.Error{
			Message: msg,
		})
	} else if err != nil {
		srv.log.Errorw("error creating kuberlogicservice", "error", err)
		return apiService.NewServiceEditBadRequest().WithPayload(
			&models.Error{
				Message: err.Error(),
			})
	}

	return apiService.NewServiceEditOK()
}
