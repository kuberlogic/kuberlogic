package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/util"
	"github.com/kuberlogic/operator/modules/operator/service-operator/util/kuberlogic"
	"github.com/pkg/errors"
)

func (srv *Service) UserCreateHandler(params apiService.UserCreateParams, principal *models.Principal) middleware.Responder {
	session, err := kuberlogic.GetSession(srv.existingService, srv.clientset, "")
	if err != nil {
		srv.log.Errorw("error generating session", "error", err)
		return util.BadRequestFromError(err)
	}

	if protected := session.GetUser().IsProtected(*params.User.Name); protected {
		e := errors.Errorf("User '%s' is protected", *params.User.Name)
		srv.log.Errorw("error creating user", "error", e)
		return util.BadRequestFromError(e)
	}

	err = session.GetUser().Create(*params.User.Name, params.User.Password)
	if err != nil {
		srv.log.Errorw("error creating user: %s", "error", err)
		return util.BadRequestFromError(err)
	}

	return apiService.NewUserCreateCreated()
}
