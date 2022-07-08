package app

import (
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiBackup "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/backup"
	cloudlinuxv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"net/http"
	"testing"
)

func TestBackupDeleteOK(t *testing.T) {
	expectedObj := &cloudlinuxv1alpha1.KuberlogicServiceBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name: "simple",
		},
		Spec: cloudlinuxv1alpha1.KuberlogicServiceBackupSpec{
			KuberlogicServiceName: "test",
		},
	}

	tc := createTestClient(expectedObj, 200, t)
	defer tc.server.Close()

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
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
