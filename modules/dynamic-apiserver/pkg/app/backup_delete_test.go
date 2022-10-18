package app

import (
	"net/http"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiBackup "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/backup"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func TestBackupDeleteOK(t *testing.T) {
	expectedObj := &v1alpha1.KuberlogicServiceBackup{
		ObjectMeta: v1.ObjectMeta{
			Name: "simple",
		},
		Spec: v1alpha1.KuberlogicServiceBackupSpec{
			KuberlogicServiceName: "test",
		},
	}

	tc := createTestClient(expectedObj, 200, t)
	defer tc.server.Close()

	srv := &handlers{
		log:        &TestLog{t: t},
		clientset:  fake.NewSimpleClientset(),
		restClient: tc.client,
	}

	backup := &models.Backup{
		ID: "simple",
	}

	params := apiBackup.BackupDeleteParams{
		HTTPRequest: &http.Request{},
		BackupID:    "simple",
	}

	checkResponse(srv.BackupDeleteHandler(params, nil), t, 200, backup)
	tc.handler.ValidateRequestCount(t, 1)
}
