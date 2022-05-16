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

func TestServiceDeleteOK(t *testing.T) {
	namespace := "-test-namespace-"
	expectedObj := &cloudlinuxv1alpha1.KuberLogicService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "simple",
			Namespace: namespace,
		},
		Spec: cloudlinuxv1alpha1.KuberLogicServiceSpec{
			Type:     "postgresql",
			Replicas: 1,
		},
	}

	tc := createTestClient(expectedObj, 200, t)
	defer tc.server.Close()

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
	}

	service := &models.Service{
		Name:     util.StrAsPointer("simple"),
		Ns:       namespace,
		Replicas: util.Int64AsPointer(1),
		Type:     util.StrAsPointer("postgresql"),
		Status:   "Unknown",
	}

	params := apiService.ServiceDeleteParams{
		HTTPRequest: &http.Request{},
		ServiceID:   "-test-namespace-:service",
	}

	checkResponse(srv.ServiceDeleteHandler(params), t, 200, service)
	tc.handler.ValidateRequestCount(t, 2) // get and delete request under the hood
}

func TestServiceDeleteIncorrectID(t *testing.T) {
	namespace := "-test-namespace-"
	expectedObj := &cloudlinuxv1alpha1.KuberLogicService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "simple",
			Namespace: namespace,
		},
		Spec: cloudlinuxv1alpha1.KuberLogicServiceSpec{
			Type:     "postgresql",
			Replicas: 1,
		},
	}

	tc := createTestClient(expectedObj, 200, t)
	defer tc.server.Close()

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
	}

	params := apiService.ServiceDeleteParams{
		HTTPRequest: &http.Request{},
		ServiceID:   "-test-namespace-service",
	}

	checkResponse(srv.ServiceDeleteHandler(params), t, 400, &models.Error{
		Message: "incorrect service id",
	})
	tc.handler.ValidateRequestCount(t, 0) // none requests are passed
}
