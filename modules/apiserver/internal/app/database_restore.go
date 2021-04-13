package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/util"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
)

func (srv *Service) DatabaseRestoreHandler(params apiService.DatabaseRestoreParams, principal *models.Principal) middleware.Responder {
	service := params.HTTPRequest.Context().Value("service").(*kuberlogicv1.KuberLogicService)
	ns, name := service.Namespace, service.Name

	srv.log.Debugw("attempting to create a restore backup resource", "namespace", ns, "name", name)
	err := util.CreateBackupRestoreResource(srv.kuberlogicClient, ns, name, *params.RestoreItem.Key, *params.RestoreItem.Database)
	if err != nil {
		srv.log.Errorw("error creating a backup restore resource",
			"namespace", ns, "name", name, "error", err)
		return util.BadRequestFromError(err)
	}

	return apiService.NewDatabaseRestoreOK()
}
