package app

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/util"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	operator "github.com/kuberlogic/operator/modules/operator/service-operator"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// set this string to a required security grant for this action
const backupListSecGrant = "service:backup:list"

func (srv *Service) BackupListHandler(params apiService.BackupListParams, principal *models.Principal) middleware.Responder {

	ns, name, err := util.SplitID(params.ServiceID)
	if err != nil {
		return util.BadRequestFromError(err)
	}

	if authorized, err := srv.authProvider.Authorize(principal.Token, backupListSecGrant, params.ServiceID); err != nil {
		srv.log.Errorf("error checking authorization: %s ", err.Error())
		resp := apiService.NewBackupListBadRequest()
		return resp
	} else if !authorized {
		resp := apiService.NewBackupListForbidden()
		return resp
	}

	srv.log.Debugf("attempting to get a backup config %s/%s", ns, name)
	secret, err := srv.clientset.CoreV1().
		Secrets(ns).
		Get(context.TODO(), name, v1.GetOptions{})
	if errors.IsNotFound(err) {
		srv.log.Errorf("backup config %s/%s does not exist: %s", ns, name, err.Error())
		return util.BadRequestFromError(err)
	} else if err != nil {
		srv.log.Errorf("failed to get a backup config %s/%s: %s", ns, name, err.Error())
		return util.BadRequestFromError(err)
	}

	// check cluster is exists
	item := &kuberlogicv1.KuberLogicService{}
	err = srv.cmClient.Get().
		Namespace(ns).
		Resource("kuberlogicservices").
		Name(name).
		Do(context.TODO()).
		Into(item)
	if err != nil {
		srv.log.Errorf("couldn't find KuberLogicService resource in cluster: %s", err.Error())
		return util.BadRequestFromError(err)
	}
	op, err := operator.GetOperator(item.Spec.Type)
	if err != nil {
		srv.log.Errorf("Could not define the base operator: %s", err)
		return util.BadRequestFromError(err)
	}

	model := util.BackupConfigResourceToModel(secret)
	mySession := session.Must(session.NewSession(
		&aws.Config{
			Endpoint: model.Endpoint,
			// region just a stub -> for s3 region is no needed, but required for sdk
			Region: aws.String("us-west-2"),
			Credentials: credentials.NewStaticCredentials(
				*model.AwsAccessKeyID,
				*model.AwsSecretAccessKey,
				""),
			S3ForcePathStyle: aws.Bool(true),
		},
	))

	prefix := fmt.Sprintf("%s/%s/logical_backups/", item.Spec.Type, op.Name(item))
	out, err := s3.New(mySession).ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: model.Bucket,
		Prefix: &prefix,
	})
	if err != nil {
		srv.log.Errorf("failed to get a backups %s/%s: %s", ns, name, err.Error())
		return apiService.NewBackupListServiceUnavailable().WithPayload(&models.Error{
			Message: err.Error(),
		})
	}
	var payload []*models.Backup
	for _, item := range out.Contents {
		dt := strfmt.DateTime(*item.LastModified)

		key := fmt.Sprintf("s3://%s/%s", *model.Bucket, *item.Key)
		payload = append(payload, &models.Backup{
			Key:          &key,
			Size:         item.Size,
			LastModified: &dt,
		})
	}

	return apiService.NewBackupListOK().WithPayload(payload)
}
