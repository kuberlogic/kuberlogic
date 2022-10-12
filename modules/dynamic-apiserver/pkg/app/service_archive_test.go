package app

import (
	"net/http"
	"testing"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/config"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	cloudlinuxv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestServiceArchiveSimple(t *testing.T) {
	serviceID := "archived_service"
	expectedObj := &cloudlinuxv1alpha1.KuberLogicService{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceID,
		},
		Spec: cloudlinuxv1alpha1.KuberLogicServiceSpec{
			Type:     "docker-compose",
			Replicas: 1,
		},
	}

	tc := createTestClient(expectedObj, 200, t)
	defer tc.server.Close()

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
		config: &config.Config{
			Domain: "example.com",
		},
	}

	service := &models.Service{
		ID:       util.StrAsPointer(serviceID),
		Replicas: util.Int64AsPointer(1),
		Type:     util.StrAsPointer("docker-compose"),
	}

	params := apiService.ServiceAddParams{
		HTTPRequest: &http.Request{},
		ServiceItem: service,
	}
	archiveParams := apiService.ServiceArchiveParams{
		HTTPRequest: &http.Request{},
		ServiceID:   serviceID,
	}

	checkResponse(srv.ServiceAddHandler(params, nil), t, 201, service)
	checkResponse(srv.ServiceArchiveHandler(archiveParams, nil), t, 200, struct{}{})
	tc.handler.ValidateRequestCount(t, 2)
}
