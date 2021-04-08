package app

import (
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/util"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// curl -v -H Content-Type:application/json -H "Authorization: Bearer" -X POST localhost:8001/api/v1/services/<service-id>/backup-config -d '{"aws_access_key_id":"SJ3MEX4WE7G2A5JLHJQC","aws_secret_access_key":"hTXfI4Gbv0SPSWGhnWQrINg6TPcWCCvLcB2DRFmp+Ok","bucket":"cloudmanaged","endpoint":"https://fra1.digitaloceanspaces.com","schedule":"* 1 * * *","type":"s3","enabled":false}'
func (srv *Service) BackupConfigCreateHandler(params apiService.BackupConfigCreateParams, principal *models.Principal) middleware.Responder {
	ns, name := srv.existingService.Namespace, srv.existingService.Name

	// Create secret
	secretResource := util.BackupConfigModelToResource(params.BackupConfig)
	secretResource.ObjectMeta = v1.ObjectMeta{
		Name:      name,
		Namespace: ns,
	}

	srv.log.Debugw("attempting to create a backup config", "namespace", ns, "name", name)
	_, err := srv.clientset.CoreV1().
		Secrets(ns).
		Create(context.TODO(), secretResource, v1.CreateOptions{})
	if err != nil {
		srv.log.Errorw("failed to create a backup config",
			"namespace", ns, "name", name, "error", err)
		return util.BadRequestFromError(err)
	}

	if *params.BackupConfig.Enabled {
		srv.log.Debugw("attempting to create a backup resource",
			"namespace", ns, "name", name)
		err = util.CreateBackupResource(srv.kuberlogicClient, ns, name, *params.BackupConfig.Schedule)
		if err != nil {
			srv.log.Errorw("error creating a backup resource",
				"namespace", ns, "name", name, "error", err)
			return util.BadRequestFromError(err)
		}
	}

	return &apiService.BackupConfigCreateCreated{
		Payload: util.BackupConfigResourceToModel(secretResource),
	}
}
