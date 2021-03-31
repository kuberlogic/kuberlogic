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
)

func (srv *Service) UserListHandler(params apiService.UserListParams, principal *models.Principal) middleware.Responder {
	// validate path parameter
	ns, name, err := util.SplitID(params.ServiceID)
	if err != nil {
		srv.log.Errorw("incorrect service id", "error", err)
		return util.BadRequestFromError(err)
	}

	if authorized, err := srv.authProvider.Authorize(principal.Token, security.UserListSecGrant, params.ServiceID); err != nil {
		srv.log.Errorw("error checking authorization", "error", err)
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
		srv.log.Errorw("couldn't find KuberLogicService resource in cluster", "error", err)
		return util.BadRequestFromError(err)
	}

	session, err := kuberlogic.GetSession(&item, srv.clientset, "")
	if err != nil {
		srv.log.Errorw("error generating session", "error", err)
		return util.BadRequestFromError(err)
	}

	users, err := session.GetUser().List()
	if err != nil {
		srv.log.Errorw("error receiving databases", "error", err)
		return util.BadRequestFromError(err)
	}

	var payload []*models.User
	for _, dbUser := range users {
		userName := dbUser

		if protected := session.GetUser().IsProtected(userName); !protected {
			payload = append(payload, &models.User{
				Name: &userName,
			})
		}
	}

	return apiService.NewUserListOK().WithPayload(payload)
}
