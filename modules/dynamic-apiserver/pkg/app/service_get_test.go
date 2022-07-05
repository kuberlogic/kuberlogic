package app

import (
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	cloudlinuxv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"net/http"
	"testing"
)

func TestServiceGetNotFound(t *testing.T) {
	expectedObject := &cloudlinuxv1alpha1.KuberLogicService{}

	tc := createTestClient(expectedObject, 404, t)
	defer tc.server.Close()

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
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
	expectedObject := &cloudlinuxv1alpha1.KuberLogicService{
		ObjectMeta: metav1.ObjectMeta{
			Name: "one",
		},
		Spec: cloudlinuxv1alpha1.KuberLogicServiceSpec{
			Type:       "postgresql",
			Replicas:   1,
			VolumeSize: "2Gi",
		},
	}

	tc := createTestClient(expectedObject, 200, t)
	defer tc.server.Close()

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
	}

	service := &models.Service{
		ID:         util.StrAsPointer("one"),
		Type:       util.StrAsPointer("postgresql"),
		Replicas:   util.Int64AsPointer(1),
		VolumeSize: "2Gi",
		Status:     "Unknown",
	}

	params := apiService.ServiceGetParams{
		HTTPRequest: &http.Request{},
		ServiceID:   "one",
	}

	checkResponse(srv.ServiceGetHandler(params, nil), t, 200, service)
	tc.handler.ValidateRequestCount(t, 1)
}
