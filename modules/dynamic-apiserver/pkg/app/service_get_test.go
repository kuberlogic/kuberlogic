package app

import (
	"net/http"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func TestServiceGetNotFound(t *testing.T) {
	expectedObject := &v1alpha1.KuberLogicService{}

	tc := createTestClient(expectedObject, 404, t)
	defer tc.server.Close()

	srv := &handlers{
		log:        &TestLog{t: t},
		clientset:  fake.NewSimpleClientset(),
		restClient: tc.client,
	}

	params := apiService.ServiceGetParams{
		HTTPRequest: &http.Request{},
		ServiceID:   "not-found-id",
	}

	checkResponse(srv.ServiceGetHandler(params, nil), t, 404, &models.Error{
		Message: "kuberlogic service not found: not-found-id",
	})
	tc.handler.ValidateRequestCount(t, 1)
}

func TestServiceGetSuccess(t *testing.T) {
	expectedObject := &v1alpha1.KuberLogicService{
		ObjectMeta: v1.ObjectMeta{
			Name: "one",
		},
		Spec: v1alpha1.KuberLogicServiceSpec{
			Type:     "postgresql",
			Replicas: 1,
			Limits: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("2Gi"),
			},
		},
		Status: v1alpha1.KuberLogicServiceStatus{
			Phase: "Unknown",
		},
	}

	tc := createTestClient(expectedObject, 200, t)
	defer tc.server.Close()

	srv := &handlers{
		log:        &TestLog{t: t},
		clientset:  fake.NewSimpleClientset(),
		restClient: tc.client,
	}

	service := &models.Service{
		ID:       util.StrAsPointer("one"),
		Type:     util.StrAsPointer("postgresql"),
		Replicas: util.Int64AsPointer(1),
		Limits: &models.Limits{
			Storage: "2Gi",
		},
		Status: "Unknown",
	}

	params := apiService.ServiceGetParams{
		HTTPRequest: &http.Request{},
		ServiceID:   "one",
	}

	checkResponse(srv.ServiceGetHandler(params, nil), t, 200, service)
	tc.handler.ValidateRequestCount(t, 1)
}
