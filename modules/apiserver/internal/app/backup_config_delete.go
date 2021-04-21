package app

import (
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/util"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (srv *Service) BackupConfigDeleteHandler(params apiService.BackupConfigDeleteParams, principal *models.Principal) middleware.Responder {
	service := params.HTTPRequest.Context().Value("service").(*kuberlogicv1.KuberLogicService)
	ns, name := service.Namespace, service.Name

	srv.log.Debugw("attempting to delete a backup config", "namespace", ns, "name", name)
	err := srv.clientset.CoreV1().Secrets(ns).
		Delete(context.TODO(), name, v1.DeleteOptions{})
	if errors.IsNotFound(err) {
		srv.log.Errorw("backup config does not exist",
			"namespace", ns, "name", name, "error", err)
		return &apiService.BackupConfigDeleteNotFound{}
	}
	if err != nil {
		srv.log.Errorw("error deleting backup config", "error", err)
		return util.BadRequestFromError(err)
	}

	srv.log.Debugw("attempting to delete a backup resource",
		"namespace", ns, "name", name)
	err = util.DeleteBackupResource(srv.kuberlogicClient, ns, name)
	if err != nil {
		srv.log.Errorw("error deleting backup resource", "error", err)
		return util.BadRequestFromError(err)
	}

	return &apiService.BackupConfigDeleteOK{}
}
