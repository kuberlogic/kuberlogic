package app

import (
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/util"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
)

func (srv *Service) RestoreListHandler(params apiService.RestoreListParams, principal *models.Principal) middleware.Responder {
	service := params.HTTPRequest.Context().Value("service").(*kuberlogicv1.KuberLogicService)
	ns, name := principal.Namespace, service.Name

	srv.log.Debugw("searching for service", "namespace", ns, "name", name)
	if _, found, errGet := srv.serviceStore.GetService(name, ns, params.HTTPRequest.Context()); errGet != nil {
		srv.log.Errorw("service get error", "error", errGet.Err)
		return apiService.NewRestoreListServiceUnavailable().WithPayload(&models.Error{Message: errGet.ClientMsg})
	} else if !found {
		return util.BadRequestFromError(fmt.Errorf("%s/%s service not found", ns, name))
	}

	srv.log.Debugw("service exists", "namespace", ns, "name", name)
	restores, errRestores := srv.serviceStore.GetServiceRestores(ns, name, params.HTTPRequest.Context())
	if errRestores != nil {
		return apiService.NewRestoreListServiceUnavailable().WithPayload(&models.Error{Message: errRestores.ClientMsg})
	}

	return apiService.NewRestoreListOK().WithPayload(restores)
}
