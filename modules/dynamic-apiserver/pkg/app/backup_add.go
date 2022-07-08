package app

import (
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiBackup "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/backup"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func (srv *Service) BackupAddHandler(params apiBackup.BackupAddParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	klb, err := util.BackupToKuberlogic(params.BackupItem)
	if err != nil {
		srv.log.Errorw("error converting backup to kuberlogic object", "error", err)
		return apiBackup.NewBackupAddBadRequest().WithPayload(&models.Error{
			Message: errors.Wrap(err, "error converting backup to kuberlogic object").Error(),
		})
	}

	kls := &v1alpha1.KuberLogicService{
		ObjectMeta: metav1.ObjectMeta{
			Name: klb.Spec.KuberlogicServiceName,
		},
	}
	if err := srv.kuberlogicClient.Get().
		Resource(serviceK8sResource).
		Name(kls.GetName()).
		Do(ctx).
		Error(); k8serrors.IsNotFound(err) {
		return apiBackup.NewBackupAddBadRequest().WithPayload(&models.Error{
			Message: fmt.Sprintf("service `%s` not found", kls.GetName()),
		})
	} else if err != nil {
		srv.log.Errorw("error getting kuberlogicservice for backup", "error", err)
		return apiBackup.NewBackupAddServiceUnavailable().WithPayload(&models.Error{
			Message: fmt.Sprintf("error getting coresponding service %s: %s", kls.GetName(), err),
		})
	}

	klb.SetName(fmt.Sprintf("%s-%d", kls.GetName(), time.Now().Unix()))

	if err := srv.kuberlogicClient.Post().
		Resource(backupK8sResource).
		Name(klb.GetName()).
		Body(klb).
		Do(ctx).
		Into(klb); k8serrors.IsAlreadyExists(err) {
		srv.log.Errorw("klb already exists", "name", klb.GetName())
		return apiBackup.NewBackupAddConflict()
	} else if err != nil {
		srv.log.Errorw("error creating klb", "error", err, "name", klb.GetName())
		return apiBackup.NewBackupAddServiceUnavailable().WithPayload(&models.Error{
			Message: err.Error(),
		})
	}

	created, err := util.KuberlogicToBackup(klb)
	if err != nil {
		srv.log.Errorw("error converting klb to models.Backup", "error", err)
		return apiBackup.NewBackupAddServiceUnavailable().WithPayload(&models.Error{
			Message: "error converting created backup",
		})
	}
	return apiBackup.NewBackupAddCreated().WithPayload(created)
}
