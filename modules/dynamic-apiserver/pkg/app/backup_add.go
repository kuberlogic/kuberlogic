package app

import (
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiBackup "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/backup"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
)

func (h *handlers) BackupAddHandler(params apiBackup.BackupAddParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	klb := util.BackupToKuberlogic(params.BackupItem)
	serviceName := klb.Spec.KuberlogicServiceName
	if _, err := h.Services().Get(ctx, serviceName, v1.GetOptions{}); k8serrors.IsNotFound(err) {
		return apiBackup.NewBackupAddBadRequest().WithPayload(&models.Error{
			Message: fmt.Sprintf("service `%s` not found", serviceName),
		})
	} else if err != nil {
		h.log.Errorw("error getting kuberlogicservice for backup", "error", err)
		return apiBackup.NewBackupAddServiceUnavailable().WithPayload(&models.Error{
			Message: fmt.Sprintf("error getting coresponding service %s: %s", serviceName, err),
		})
	}

	klb, err := h.Backups().CreateByServiceName(ctx, klb.Spec.KuberlogicServiceName)
	if k8serrors.IsAlreadyExists(err) {
		h.log.Errorw("klb already exists", "service", serviceName)
		return apiBackup.NewBackupAddConflict()
	} else if err != nil {
		h.log.Errorw("error creating klb", "error", err, "service", serviceName)
		return apiBackup.NewBackupAddServiceUnavailable().WithPayload(&models.Error{
			Message: err.Error(),
		})
	}
	return apiBackup.NewBackupAddCreated().WithPayload(util.KuberlogicToBackup(klb))
}
