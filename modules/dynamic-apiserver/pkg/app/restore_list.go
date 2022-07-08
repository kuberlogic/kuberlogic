package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiRestore "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/restore"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
)

func (srv *Service) RestoreListHandler(params apiRestore.RestoreListParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	klrs, err := srv.ListKuberlogicServiceRestoresByService(ctx, params.ServiceID)
	if err != nil {
		msg := "error listing restores"
		srv.log.Errorw(msg)
		return apiRestore.NewRestoreListServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}
	srv.log.Debugw("found kuberlogicservicerestores objects", "count", len(klrs.Items), "objects", klrs)

	restores := make([]*models.Restore, 0)
	for _, klr := range klrs.Items {
		b, err := util.KuberlogicToRestore(&klr)
		if err != nil {
			srv.log.Errorw("error converting klr to model", "error", err, "name", klr.GetName())
			return apiRestore.NewRestoreListServiceUnavailable().WithPayload(&models.Error{
				Message: "error converting restore object to model",
			})
		}
		restores = append(restores, b)
	}
	return apiRestore.NewRestoreListOK().WithPayload(restores)
}
