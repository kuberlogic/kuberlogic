package app

import (
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/internal/security"
	"github.com/kuberlogic/operator/modules/apiserver/util"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
)

func (srv *Service) DatabaseRestoreHandler(params apiService.DatabaseRestoreParams, principal *models.Principal) middleware.Responder {

	// validate path parameter
	ns, name, err := util.SplitID(params.ServiceID)
	if err != nil {
		srv.log.Errorw("incorrect service id", "error", err)
		return util.BadRequestFromError(err)
	}

	if authorized, err := srv.authProvider.Authorize(principal.Token, security.DatabaseRestoreSecGrant, params.ServiceID); err != nil {
		srv.log.Errorw("error checking authorization", "error", err)
		resp := apiService.NewDatabaseDeleteBadRequest()
		return resp
	} else if !authorized {
		resp := apiService.NewDatabaseDeleteForbidden()
		return resp
	}

	// check cluster is exists
	item := kuberlogicv1.KuberLogicService{}
	err = srv.cmClient.Get().
		Namespace(ns).
		Resource("kuberlogicservices").
		Name(name).
		Do(context.TODO()).
		Into(&item)
	if err != nil {
		srv.log.Errorw("couldn't find KuberLogicService resource in cluster", "error", err)
		return util.BadRequestFromError(err)
	}

	srv.log.Debugw("attempting to create a restore backup resource", "namespace", ns, "name", name)
	err = util.CreateBackupRestoreResource(srv.cmClient, ns, name, *params.RestoreItem.Key, *params.RestoreItem.Database)
	if err != nil {
		srv.log.Errorw("error creating a backup restore resource",
			"namespace", ns, "name", name, "error", err)
		return util.BadRequestFromError(err)
	}

	return apiService.NewDatabaseRestoreOK()
}
