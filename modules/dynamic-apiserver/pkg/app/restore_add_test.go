package app

import (
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiRestore "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/restore"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"net/http"
	"testing"
)

func TestRestoreAdd(t *testing.T) {
	expectedKlr := &kuberlogiccomv1alpha1.KuberlogicServiceRestore{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: kuberlogiccomv1alpha1.KuberlogicServiceRestoreSpec{
			KuberlogicServiceBackup: "test",
		},
	}

	tc := createTestClient(expectedKlr, 200, t)

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
	}

	restore := &models.Restore{
		BackupID: expectedKlr.Spec.KuberlogicServiceBackup,
	}

	params := apiRestore.RestoreAddParams{
		HTTPRequest: &http.Request{},
		RestoreItem: restore,
	}

	expectedRestore := &models.Restore{
		BackupID: expectedKlr.Spec.KuberlogicServiceBackup,
		ID:       expectedKlr.Spec.KuberlogicServiceBackup,
	}

	checkResponse(srv.RestoreAddHandler(params, nil), t, 201, expectedRestore)
	tc.handler.ValidateRequestCount(t, 2)
}
