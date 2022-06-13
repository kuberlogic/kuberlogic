package app

import (
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/util"
	cloudlinuxv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"net/http"
	"testing"
)

func TestServiceEditNotFound(t *testing.T) {
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

	tc := createTestClient(expectedObject, 404, t)
	defer tc.server.Close()

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
	}

	params := apiService.ServiceEditParams{
		HTTPRequest: &http.Request{},
		ServiceID:   "not-found-id",
		ServiceItem: &models.Service{
			ID:         util.StrAsPointer("one"),
			Type:       util.StrAsPointer("postgresql"),
			Replicas:   util.Int64AsPointer(1),
			VolumeSize: "2Gi",
			Status:     "Unknown",
		},
	}

	checkResponse(srv.ServiceEditHandler(params, nil), t, 404, &models.Error{
		Message: "kuberlogic service not found: not-found-id",
	})
	tc.handler.ValidateRequestCount(t, 1)
}

func TestServiceEditSuccess(t *testing.T) {
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

	params := apiService.ServiceEditParams{
		HTTPRequest: &http.Request{},
		ServiceID:   "one",
		ServiceItem: service,
	}

	checkResponse(srv.ServiceEditHandler(params, nil), t, 200, service)
	tc.handler.ValidateRequestCount(t, 1)
}
