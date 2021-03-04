package app

import (
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/util"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/watcher/api"
	"github.com/pkg/errors"
)

// set this string to a required security grant for this action
const databaseDeleteSecGrant = "service:database:delete"

func (srv *Service) DatabaseDeleteHandler(params apiService.DatabaseDeleteParams, principal *models.Principal) middleware.Responder {
	// validate path parameter
	ns, name, err := util.SplitID(params.ServiceID)
	if err != nil {
		srv.log.Errorf("incorrect service id: %s", err.Error())
		return util.BadRequestFromError(err)
	}

	if authorized, err := srv.authProvider.Authorize(principal.Token, databaseDeleteSecGrant, params.ServiceID); err != nil {
		srv.log.Errorf("error checking authorization: %s", err.Error())
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
		srv.log.Errorf("couldn't find KuberLogicService resource in cluster: %s", err.Error())
		return util.BadRequestFromError(err)
	}

	session, err := api.GetSession(&item, srv.clientset, "", "")
	if err != nil {
		srv.log.Errorf("error generating session: %s", err.Error())
		return util.BadRequestFromError(err)
	}

	if protected := session.GetDatabase().IsProtected(params.Database); protected {
		e := errors.Errorf("Database '%s' is protected", params.Database)
		srv.log.Errorf("error creating db: %s", e.Error())
		return util.BadRequestFromError(e)
	}

	err = session.GetDatabase().Drop(params.Database)
	if err != nil {
		srv.log.Errorf("error deleting db: %s", err.Error())
		return util.BadRequestFromError(err)
	}

	return apiService.NewDatabaseDeleteOK()
}
