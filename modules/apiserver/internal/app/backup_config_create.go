package app

import (
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/util"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// set this string to a required security grant for this action
const backupConfigCreateSecGrant = "service:backup-config:add"

// curl -v -H Content-Type:application/json -H "Authorization: Bearer" -X POST localhost:8001/api/v1/services/<service-id>/backup-config -d '{"aws_access_key_id":"SJ3MEX4WE7G2A5JLHJQC","aws_secret_access_key":"hTXfI4Gbv0SPSWGhnWQrINg6TPcWCCvLcB2DRFmp+Ok","bucket":"cloudmanaged","endpoint":"https://fra1.digitaloceanspaces.com","schedule":"* 1 * * *","type":"s3","enabled":false}'
func (srv *Service) BackupConfigCreateHandler(params apiService.BackupConfigCreateParams, principal *models.Principal) middleware.Responder {
	// validate path parameter
	ns, name, err := util.SplitID(params.ServiceID)
	if err != nil {
		srv.log.Errorf("incorrect service id: %s", err.Error())
		return util.BadRequestFromError(err)
	}

	if authorized, err := srv.authProvider.Authorize(principal.Token, backupConfigCreateSecGrant, params.ServiceID); err != nil {
		srv.log.Errorf("error checking authorization: %s", err.Error())
		resp := apiService.NewBackupConfigCreateBadRequest()
		return resp
	} else if !authorized {
		resp := apiService.NewBackupConfigCreateForbidden()
		return resp
	}

	// check cluster is exists
	item := kuberlogicv1.KuberLogicService{}
	err = srv.cmClient.Get().
		Namespace(ns).
		Resource("kuberlogicservices").
		Name(name).
		Do(context.TODO()).
		Into(&item)
	if err != nil {
		srv.log.Errorf("couldn't find KuberLogicService resource in cluster: %s", err.Error())
		return util.BadRequestFromError(err)
	}

	// Create secret
	secretResource := util.BackupConfigModelToResource(params.BackupConfig)
	secretResource.ObjectMeta = v1.ObjectMeta{
		Name:      name,
		Namespace: ns,
	}

	srv.log.Debugf("attempting to create a backup config %s/%s", ns, name)
	_, err = srv.clientset.CoreV1().
		Secrets(ns).
		Create(context.TODO(), secretResource, v1.CreateOptions{})
	if err != nil {
		srv.log.Errorf("failed to create a backup config %s/%s: %s", ns, name, err.Error())
		return util.BadRequestFromError(err)
	}

	if *params.BackupConfig.Enabled {
		srv.log.Debugf("attempting to create a backup resource %s/%s", ns, name)
		err = util.CreateBackupResource(srv.cmClient, ns, name, *params.BackupConfig.Schedule)
		if err != nil {
			srv.log.Errorf("error creating a backup resource %s/%s: %s", ns, name, err.Error())
			return util.BadRequestFromError(err)
		}
	}

	return &apiService.BackupConfigCreateCreated{
		Payload: util.BackupConfigResourceToModel(secretResource),
	}
}
