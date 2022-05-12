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

func TestServiceListEmpty(t *testing.T) {
	expectedObjects := &cloudlinuxv1alpha1.KuberLogicServiceList{
		Items: []cloudlinuxv1alpha1.KuberLogicService{},
	}

	tc := createTestClient(expectedObjects, 200, t)
	defer tc.server.Close()

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
	}

	params := apiService.ServiceListParams{
		HTTPRequest: &http.Request{},
	}

	checkResponse(srv.ServiceListHandler(params), t, 200, models.Services{})
	tc.handler.ValidateRequestCount(t, 1)
}

func TestServiceListMany(t *testing.T) {
	namespace := "-test-namespace-"

	expectedObjects := &cloudlinuxv1alpha1.KuberLogicServiceList{
		Items: []cloudlinuxv1alpha1.KuberLogicService{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "one",
					Namespace: namespace,
				},
				Spec: cloudlinuxv1alpha1.KuberLogicServiceSpec{
					Type:     "postgresql",
					Replicas: 1,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "two",
					Namespace: namespace,
				},
				Spec: cloudlinuxv1alpha1.KuberLogicServiceSpec{
					Type:     "mysql",
					Replicas: 2,
				},
			},
		},
	}

	tc := createTestClient(expectedObjects, 200, t)
	defer tc.server.Close()

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
	}

	services := models.Services{
		{
			Name:     util.StrAsPointer("one"),
			Ns:       namespace,
			Type:     util.StrAsPointer("postgresql"),
			Replicas: util.Int64AsPointer(1),
			Status:   "Unknown",
		},
		{
			Name:     util.StrAsPointer("two"),
			Ns:       namespace,
			Type:     util.StrAsPointer("mysql"),
			Replicas: util.Int64AsPointer(2),
			Status:   "Unknown",
		},
	}

	params := apiService.ServiceListParams{
		HTTPRequest: &http.Request{},
		Namespace:   namespace,
	}

	checkResponse(srv.ServiceListHandler(params), t, 200, services)
	tc.handler.ValidateRequestCount(t, 1)
}
