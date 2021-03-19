package app

import (
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/util"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/service-operator/util/kuberlogic"
)

// set this string to a required security grant for this action
const databaseListSecGrant = "service:database:list"

func (srv *Service) DatabaseListHandler(params apiService.DatabaseListParams, principal *models.Principal) middleware.Responder {
	// validate path parameter
	ns, name, err := util.SplitID(params.ServiceID)
	if err != nil {
		srv.log.Errorf("incorrect service id: %s", err.Error())
		return util.BadRequestFromError(err)
	}

	if authorized, err := srv.authProvider.Authorize(principal.Token, databaseListSecGrant, params.ServiceID); err != nil {
		srv.log.Errorf("error checking authorization: %s", err.Error())
		resp := apiService.NewDatabaseListBadRequest()
		return resp
	} else if !authorized {
		resp := apiService.NewDatabaseListForbidden()
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

	session, err := kuberlogic.GetSession(&item, srv.clientset, "")
	if err != nil {
		srv.log.Errorf("error generating session: %s", err.Error())
		return util.BadRequestFromError(err)
	}

	databases, err := session.GetDatabase().List()
	if err != nil {
		srv.log.Errorf("error receiving databases: %s", err.Error())
		return util.BadRequestFromError(err)
	}

	var payload []*models.Database
	for _, dbName := range databases {
		db := dbName
		if protected := session.GetDatabase().IsProtected(db); !protected {
			payload = append(payload, &models.Database{
				Name: &db,
			})
		}
	}

	return apiService.NewDatabaseListOK().WithPayload(payload)
}
