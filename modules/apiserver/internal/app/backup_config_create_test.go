package app

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/restapi/operations/service"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"net/http"
	"strings"
	"testing"
)

func TestBackupConfigCreateEnabled(t *testing.T) {
	namespace := "-test-namespace-"
	expectedObj := &kuberlogicv1.KuberLogicBackupSchedule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: namespace,
		},
		Spec: kuberlogicv1.KuberLogicBackupScheduleSpec{
			Type:        "s3",
			Schedule:    "* 1 * * *",
			ClusterName: "test",
			SecretName:  "test",
		},
	}

	tc := createTestClient(expectedObj, 200, t)
	defer tc.server.Close()

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
	}

	service := &kuberlogicv1.KuberLogicService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: namespace,
		},
	}

	backupConfig := &models.BackupConfig{
		AwsAccessKeyID:     aws.String("-test-access-key-"),
		AwsSecretAccessKey: aws.String("secret-key"),
		Bucket:             aws.String("test-bucket"),
		Enabled:            aws.Bool(true),
		Endpoint:           aws.String("https://endpoint.com"),
		Region:             aws.String("-test-region-"),
		Schedule:           aws.String("* 1 * * *"),
	}

	req := &http.Request{}
	req = req.WithContext(context.WithValue(req.Context(), "service", service))
	params := apiService.BackupConfigCreateParams{
		HTTPRequest:  req,
		BackupConfig: backupConfig,
	}

	checkResponse(srv.BackupConfigCreateHandler(params, &models.Principal{
		Namespace: namespace,
	}), t, 201, backupConfig)
	tc.handler.ValidateRequestCount(t, 1)
}

func TestBackupConfigCreateFailed(t *testing.T) {
	namespace := "-test-namespace-"
	expectedObj := &kuberlogicv1.KuberLogicBackupSchedule{}

	tc := createTestClient(expectedObj, 404, t)
	defer tc.server.Close()

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
	}

	service := &kuberlogicv1.KuberLogicService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: namespace,
		},
	}

	backupConfig := &models.BackupConfig{
		AwsAccessKeyID:     aws.String("-test-access-key-"),
		AwsSecretAccessKey: aws.String("secret-key"),
		Bucket:             aws.String("test-bucket"),
		Enabled:            aws.Bool(true),
		Endpoint:           aws.String("https://endpoint.com"),
		Region:             aws.String("-test-region-"),
		Schedule:           aws.String("* 1 * * *"),
	}

	req := &http.Request{}
	req = req.WithContext(context.WithValue(req.Context(), "service", service))
	params := apiService.BackupConfigCreateParams{
		HTTPRequest:  req,
		BackupConfig: backupConfig,
	}
	checkMessage := func(payload interface{}) {
		err := payload.(*models.Error)
		expected := "error creating a backup resource: cannot create kuberlogicbackupschedule resource"
		if !strings.HasPrefix(err.Message, expected) {
			t.Error("error does not equal: actual vs expected",
				payload, expected)
		}
	}

	checkResponse(srv.BackupConfigCreateHandler(params, &models.Principal{
		Namespace: namespace,
	}), t, 400, checkMessage)
}

func TestBackupConfigCreateNotEnabled(t *testing.T) {
	namespace := "-test-namespace-"
	expectedObj := &kuberlogicv1.KuberLogicBackupSchedule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: namespace,
		},
		Spec: kuberlogicv1.KuberLogicBackupScheduleSpec{
			Type:        "s3",
			Schedule:    "* 1 * * *",
			ClusterName: "test",
			SecretName:  "test",
		},
	}
	tc := createTestClient(expectedObj, 200, t)
	defer tc.server.Close()

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
	}

	service := &kuberlogicv1.KuberLogicService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: namespace,
		},
	}

	backupConfig := &models.BackupConfig{
		AwsAccessKeyID:     aws.String("-test-access-key-"),
		AwsSecretAccessKey: aws.String("secret-key"),
		Bucket:             aws.String("test-bucket"),
		Enabled:            aws.Bool(false),
		Endpoint:           aws.String("https://endpoint.com"),
		Region:             aws.String("-test-region-"),
		Schedule:           aws.String("* 1 * * *"),
	}

	req := &http.Request{}
	req = req.WithContext(context.WithValue(req.Context(), "service", service))
	params := apiService.BackupConfigCreateParams{
		HTTPRequest:  req,
		BackupConfig: backupConfig,
	}

	checkResponse(srv.BackupConfigCreateHandler(params, &models.Principal{
		Namespace: namespace,
	}), t, 201, backupConfig)
	// if config is disabled no need to create KuberLogicBackupSchedule resource
	tc.handler.ValidateRequestCount(t, 0)
}
