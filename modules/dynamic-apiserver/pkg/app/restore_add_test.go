package app

import (
	"net/http"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiRestore "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/restore"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func TestRestoreAdd(t *testing.T) {
	expectedKlr := &v1alpha1.KuberlogicServiceRestore{
		ObjectMeta: v1.ObjectMeta{
			Name: "test",
		},
		Spec: v1alpha1.KuberlogicServiceRestoreSpec{
			KuberlogicServiceBackup: "test",
		},
	}

	tc := createTestClient(expectedKlr, 200, t)

	srv := New(nil, fake.NewSimpleClientset(), tc.client, &TestLog{t: t})

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
