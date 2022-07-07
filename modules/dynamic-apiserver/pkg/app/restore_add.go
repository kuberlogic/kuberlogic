package app

import (
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiRestore "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/restore"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (srv *Service) RestoreAddHandler(params apiRestore.RestoreAddParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	klb := &v1alpha1.KuberlogicServiceBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name: params.RestoreItem.BackupID,
		},
	}
	if err := srv.kuberlogicClient.Get().
		Resource(backupK8sResource).
		Name(klb.GetName()).
		Do(ctx).
		Error(); k8serrors.IsNotFound(err) {
		return apiRestore.NewRestoreAddBadRequest().WithPayload(&models.Error{
			Message: fmt.Sprintf("backup `%s` not found", klb.GetName()),
		})
	} else if err != nil {
		srv.log.Errorw("error getting kuberlogicservicebackup for restore", "error", err)
		return apiRestore.NewRestoreAddServiceUnavailable().WithPayload(&models.Error{
			Message: fmt.Sprintf("error getting coresponding backup %s: %s", klb.GetName(), err),
		})
	}

	klr, err := util.RestoreToKuberlogic(params.RestoreItem, klb)
	if err != nil {
		srv.log.Errorw("error converting restore to kuberlogic object", "error", err)
		return apiRestore.NewRestoreAddBadRequest().WithPayload(&models.Error{
			Message: errors.Wrap(err, "error converting backup to kuberlogic object").Error(),
		})
	}
	klr.SetName(klb.GetName())

	if err := srv.kuberlogicClient.Post().
		Resource(restoreK8sResource).
		Name(klr.GetName()).
		Body(klr).
		Do(ctx).
		Into(klr); k8serrors.IsAlreadyExists(err) {
		srv.log.Errorw("klr already exists", "name", klr.GetName())
		return apiRestore.NewRestoreAddConflict()
	} else if err != nil {
		srv.log.Errorw("error creating klr", "error", err, "name", klr.GetName())
		return apiRestore.NewRestoreAddServiceUnavailable().WithPayload(&models.Error{
			Message: err.Error(),
		})
	}

	created, err := util.KuberlogicToRestore(klr)
	if err != nil {
		srv.log.Errorw("error converting klr to models.Restore", "error", err)
		return apiRestore.NewRestoreAddServiceUnavailable().WithPayload(&models.Error{
			Message: "error converting created restore",
		})
	}
	return apiRestore.NewRestoreAddCreated().WithPayload(created)
}
