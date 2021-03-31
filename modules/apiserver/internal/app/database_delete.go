package app

import (
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/internal/security"
	"github.com/kuberlogic/operator/modules/apiserver/util"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/service-operator/util/kuberlogic"
	"github.com/pkg/errors"
)

func (srv *Service) DatabaseDeleteHandler(params apiService.DatabaseDeleteParams, principal *models.Principal) middleware.Responder {
	// validate path parameter
	ns, name, err := util.SplitID(params.ServiceID)
	if err != nil {
		srv.log.Errorw("incorrect service id", "error", err)
		return util.BadRequestFromError(err)
	}

	if authorized, err := srv.authProvider.Authorize(principal.Token, security.DatabaseDeleteSecGrant, params.ServiceID); err != nil {
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

	session, err := kuberlogic.GetSession(&item, srv.clientset, "")
	if err != nil {
		srv.log.Errorw("error generating session: %s", "error", err)
		return util.BadRequestFromError(err)
	}

	if protected := session.GetDatabase().IsProtected(params.Database); protected {
		e := errors.Errorf("Database '%s' is protected", params.Database)
		srv.log.Errorw("error creating db", "error", e)
		return util.BadRequestFromError(e)
	}

	err = session.GetDatabase().Drop(params.Database)
	if err != nil {
		srv.log.Errorw("error deleting db", "error", err.Error())
		return util.BadRequestFromError(err)
	}

	return apiService.NewDatabaseDeleteOK()
}
