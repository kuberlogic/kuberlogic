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

func TestRestoreDeleteOK(t *testing.T) {
	expectedObj := &v1alpha1.KuberlogicServiceRestore{
		ObjectMeta: v1.ObjectMeta{
			Name: "simple",
		},
		Spec: v1alpha1.KuberlogicServiceRestoreSpec{
			KuberlogicServiceBackup: "test",
		},
	}

	tc := createTestClient(expectedObj, 200, t)
	defer tc.server.Close()

	srv := &handlers{
		log:        &TestLog{t: t},
		clientset:  fake.NewSimpleClientset(),
		restClient: tc.client,
	}

	restore := &models.Restore{
		ID: "simple",
	}

	params := apiRestore.RestoreDeleteParams{
		HTTPRequest: &http.Request{},
		RestoreID:   "simple",
	}

	checkResponse(srv.RestoreDeleteHandler(params, nil), t, 200, restore)
	tc.handler.ValidateRequestCount(t, 1)
}
