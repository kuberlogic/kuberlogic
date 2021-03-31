package app

import (
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/internal/security"
	"github.com/kuberlogic/operator/modules/apiserver/util"
)

func (srv *Service) RestoreListHandler(params apiService.RestoreListParams, principal *models.Principal) middleware.Responder {

	ns, name, err := util.SplitID(params.ServiceID)
	if err != nil {
		return util.BadRequestFromError(err)
	}

	if authorized, err := srv.authProvider.Authorize(principal.Token, security.RestoreListSecGrant, params.ServiceID); err != nil {
		srv.log.Errorw("error checking authorization ", "error", err)
		resp := apiService.NewBackupListBadRequest()
		return resp
	} else if !authorized {
		resp := apiService.NewBackupListForbidden()
		return resp
	}

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
