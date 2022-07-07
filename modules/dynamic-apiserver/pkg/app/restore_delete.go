package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiRestore "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/restore"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (srv *Service) RestoreDeleteHandler(params apiRestore.RestoreDeleteParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	if err := srv.kuberlogicClient.Delete().
		Resource(restoreK8sResource).
		Name(params.RestoreID).
		Do(ctx).
		Error(); errors.IsNotFound(err) {
		return apiRestore.NewRestoreDeleteNotFound()
	} else if err != nil {
		srv.log.Errorw("error deleting klr", "error", err, "name", params.RestoreID)
		return apiRestore.NewRestoreDeleteServiceUnavailable().WithPayload(&models.Error{
			Message: "error deleting restore",
		})
	}
	return apiRestore.NewRestoreDeleteOK()
}
