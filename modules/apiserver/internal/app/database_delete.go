package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/util"
	"github.com/kuberlogic/operator/modules/operator/service-operator/util/kuberlogic"
	"github.com/pkg/errors"
)

func (srv *Service) DatabaseDeleteHandler(params apiService.DatabaseDeleteParams, principal *models.Principal) middleware.Responder {
	session, err := kuberlogic.GetSession(srv.existingService, srv.clientset, "")
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
