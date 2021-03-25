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
		srv.log.Errorw("error checking authorization", "error", err)
		resp := apiService.NewBackupListBadRequest()
		return resp
	} else if !authorized {
		resp := apiService.NewBackupListForbidden()
		return resp
	}

	srv.log.Debugw("attempting to get a backup config",
		"namespace", ns, "name", name)
	secret, err := srv.clientset.CoreV1().
		Secrets(ns).
		Get(context.TODO(), name, v1.GetOptions{})
	if errors.IsNotFound(err) {
		srv.log.Errorw("backup config does not exist",
			"namespace", ns, "name", name, "error", err)
		return util.BadRequestFromError(err)
	} else if err != nil {
		srv.log.Errorw("failed to get a backup config",
			"namespace", ns, "name", name, "error", err)
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
		srv.log.Errorw("couldn't find KuberLogicService resource in cluster", "error", err)
		return util.BadRequestFromError(err)
	}
	op, err := operator.GetOperator(item.Spec.Type)
	if err != nil {
		srv.log.Errorw("Could not define the base operator", "error", err)
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
		srv.log.Errorw("failed to get a backups",
			"namespace", ns, "name", name, "error", err)
		return apiService.NewBackupListServiceUnavailable().WithPayload(&models.Error{
			Message: err.Error(),
		})
	}
	var payload []*models.Backup
	for _, item := range out.Contents {
		dt := strfmt.DateTime(*item.LastModified)

		key := fmt.Sprintf("s3://%s/%s", *model.Bucket, *item.Key)
		payload = append(payload, &models.Backup{
			File:         &key,
			Size:         item.Size,
			LastModified: &dt,
		})
	}

	return apiService.NewBackupListOK().WithPayload(payload)
}
