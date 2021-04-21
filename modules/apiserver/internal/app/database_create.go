package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/util"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/service-operator/util/kuberlogic"
	"github.com/pkg/errors"
)

func (srv *Service) DatabaseCreateHandler(params apiService.DatabaseCreateParams, principal *models.Principal) middleware.Responder {
	service := params.HTTPRequest.Context().Value("service").(*kuberlogicv1.KuberLogicService)
	session, err := kuberlogic.GetSession(service, srv.clientset, "")
	if err != nil {
		srv.log.Errorw("error generating session", "error", err)
		return util.BadRequestFromError(err)
	}

	if protected := session.GetDatabase().IsProtected(*params.Database.Name); protected {
		e := errors.Errorf("Database '%s' is protected", *params.Database.Name)
		srv.log.Errorw("error creating db", "error", e)
		return util.BadRequestFromError(e)
	}

	err = session.GetDatabase().Create(*params.Database.Name)
	if err != nil {
		srv.log.Errorw("error creating db", "error", err)
		return util.BadRequestFromError(err)
	}

	return apiService.NewDatabaseCreateCreated()
}
