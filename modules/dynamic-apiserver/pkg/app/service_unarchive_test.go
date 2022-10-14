package app

import (
	"net/http"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/config"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func TestServiceUnarchive(t *testing.T) {
	t.Skip("Skipping test due to fail gorotine under the UnarchiveKuberlogicService")

	serviceID := "archived_service"
	expectedObj := &v1alpha1.KuberLogicService{
		ObjectMeta: v1.ObjectMeta{
			Name: serviceID,
		},
		Spec: v1alpha1.KuberLogicServiceSpec{
			Type:     "docker-compose",
			Replicas: 1,
			Archived: true,
		},
	}
	expectedObj.MarkArchived()

	tc := createTestClient(expectedObj, 200, t)
	defer tc.server.Close()

	srv := &handlers{
		log:        &TestLog{t: t},
		clientset:  fake.NewSimpleClientset(),
		restClient: tc.client,
		config: &config.Config{
			Domain: "example.com",
		},
	}

	unarchiveParams := apiService.ServiceUnarchiveParams{
		HTTPRequest: &http.Request{},
		ServiceID:   serviceID,
	}

	checkResponse(srv.ServiceUnarchiveHandler(unarchiveParams, nil), t, 200, struct{}{})
	tc.handler.ValidateRequestCount(t, 1)
}
