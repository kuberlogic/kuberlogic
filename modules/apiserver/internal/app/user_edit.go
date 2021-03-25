package app

import (
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/util"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/service-operator/util/kuberlogic"
	"github.com/pkg/errors"
)

// set this string to a required security grant for this action
const userEditSecGrant = "service:user:edit"

func (srv *Service) UserEditHandler(params apiService.UserEditParams, principal *models.Principal) middleware.Responder {
	// validate path parameter
	ns, name, err := util.SplitID(params.ServiceID)
	if err != nil {
		srv.log.Errorw("incorrect service id", "error", err)
		return util.BadRequestFromError(err)
	}

	if authorized, err := srv.authProvider.Authorize(principal.Token, userEditSecGrant, params.ServiceID); err != nil {
		srv.log.Errorw("error checking authorization", "error", err)
		resp := apiService.NewUserEditBadRequest()
		return resp
	} else if !authorized {
		resp := apiService.NewUserEditForbidden()
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

	if protected := session.GetUser().IsProtected(*params.User.Name); protected {
		e := errors.Errorf("User '%s' is protected", *params.User.Name)
		srv.log.Errorw("error creating user", "error", e)
		return util.BadRequestFromError(e)
	}

	err = session.GetUser().Edit(*params.User.Name, params.User.Password)
	if err != nil {
		srv.log.Errorw("error editing user", "error", err)
		return util.BadRequestFromError(err)
	}

	return apiService.NewUserEditOK()
}
