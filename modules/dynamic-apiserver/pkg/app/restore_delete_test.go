package app

import (
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiRestore "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/restore"
	cloudlinuxv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"net/http"
	"testing"
)

func TestRestoreDeleteOK(t *testing.T) {
	expectedObj := &cloudlinuxv1alpha1.KuberlogicServiceRestore{
		ObjectMeta: metav1.ObjectMeta{
			Name: "simple",
		},
		Spec: cloudlinuxv1alpha1.KuberlogicServiceRestoreSpec{
			KuberlogicServiceBackup: "test",
		},
	}

	tc := createTestClient(expectedObj, 200, t)
	defer tc.server.Close()

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
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
