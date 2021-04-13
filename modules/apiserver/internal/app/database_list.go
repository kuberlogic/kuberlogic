package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/util"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/service-operator/util/kuberlogic"
)

func (srv *Service) DatabaseListHandler(params apiService.DatabaseListParams, principal *models.Principal) middleware.Responder {
	service := params.HTTPRequest.Context().Value("service").(*kuberlogicv1.KuberLogicService)
	session, err := kuberlogic.GetSession(service, srv.clientset, "")
	if err != nil {
		srv.log.Errorw("error generating session", "error", err)
		return util.BadRequestFromError(err)
	}

	databases, err := session.GetDatabase().List()
	if err != nil {
		srv.log.Errorw("error receiving databases", "error", err)
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
