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

func (srv *Service) ServiceUnarchiveHandler(params apiService.ServiceUnarchiveParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()
	// Check if service exists first
	kls := new(kuberlogiccomv1alpha1.KuberLogicService)
	err := srv.kuberlogicClient.Get().
		Resource(serviceK8sResource).
		Name(params.ServiceID).
		Do(ctx).
		Into(kls)
	if err != nil && util.CheckStatus(err, v1.StatusReasonNotFound) {
		msg := fmt.Sprintf("kuberlogic service not found: %s", params.ServiceID)
		srv.log.Errorw(msg, "error", err)
		return apiService.NewServiceUnarchiveNotFound().WithPayload(&models.Error{
			Message: msg,
		})
	} else if err != nil {
		msg := "error finding service"
		srv.log.Errorw(msg, "error", err)
		return apiService.NewServiceUnarchiveServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}
	if !kls.Archived() {
		msg := fmt.Sprintf("service is not in archive state: %s", kls.GetName())
		srv.log.Errorw(msg)
		return apiService.NewServiceUnarchiveServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}

	// Unarchive the service in background
	go func() {
		if err := srv.UnarchiveKuberlogicService(kls); err != nil {
			srv.log.Errorw("error unarchiving service", "error", err)
		}
	}()
	return apiService.NewServiceUnarchiveOK()
}
