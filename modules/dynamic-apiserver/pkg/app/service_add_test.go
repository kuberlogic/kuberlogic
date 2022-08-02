package app

import (
	"encoding/json"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	cloudlinuxv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"net/http"
	"testing"
)

func TestServiceAddSimple(t *testing.T) {
	expectedObj := &cloudlinuxv1alpha1.KuberLogicService{
		ObjectMeta: metav1.ObjectMeta{
			Name: "simple",
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
		ID:       util.StrAsPointer("simple"),
		Replicas: util.Int64AsPointer(1),
		Type:     util.StrAsPointer("postgresql"),
	}

	params := apiService.ServiceAddParams{
		HTTPRequest: &http.Request{},
		ServiceItem: service,
	}

	checkResponse(srv.ServiceAddHandler(params, nil), t, 201, service)
	tc.handler.ValidateRequestCount(t, 1)
}

func TestServiceAddExtended(t *testing.T) {
	expectedObj := &cloudlinuxv1alpha1.KuberLogicService{
		ObjectMeta: metav1.ObjectMeta{
			Name: "extended",
		},
		Spec: cloudlinuxv1alpha1.KuberLogicServiceSpec{
			Type:     "postgresql",
			Replicas: 1,
			Limits: v1.ResourceList{
				"cpu":     resource.MustParse("10"),
				"memory":  resource.MustParse("500"),
				"storage": resource.MustParse("100Gi"),
			},
			Version:        "13",
			BackupSchedule: "*/10 * * * *",
		},
	}
	expectedObj.MarkReady("Ready")

	tc := createTestClient(expectedObj, 200, t)
	defer tc.server.Close()

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
	}

	service := &models.Service{
		ID:       util.StrAsPointer("extended"),
		Replicas: util.Int64AsPointer(1),
		Type:     util.StrAsPointer("postgresql"),
		Limits: &models.Limits{
			CPU:        "10",
			Memory:     "500",
			VolumeSize: "100Gi",
		},
		Version:        "13",
		Status:         "Ready",
		BackupSchedule: "*/10 * * * *",
	}

	params := apiService.ServiceAddParams{
		HTTPRequest: &http.Request{},
		ServiceItem: service,
	}

	checkResponse(srv.ServiceAddHandler(params, nil), t, 201, service)
	tc.handler.ValidateRequestCount(t, 1)
}

func TestServiceAddAdvanced(t *testing.T) {
	expectedObj := &cloudlinuxv1alpha1.KuberLogicService{
		ObjectMeta: metav1.ObjectMeta{
			Name: "advanced",
			Labels: map[string]string{
				"subscription-id": "some-kind-of-subscription-id",
			},
		},
		Spec: cloudlinuxv1alpha1.KuberLogicServiceSpec{
			Type: "postgresql",
		},
	}

	advanced := map[string]interface{}{
		"one": "1",
		"two": float64(2),
		"free": map[string]interface{}{
			"bool": true,
		},
	}

	bytes, _ := json.Marshal(advanced)
	expectedObj.Spec.Advanced.Raw = bytes

	tc := createTestClient(expectedObj, 200, t)
	defer tc.server.Close()

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
	}

	service := &models.Service{
		ID:           util.StrAsPointer("advanced"),
		Type:         util.StrAsPointer("postgresql"),
		Replicas:     util.Int64AsPointer(0),
		Advanced:     advanced,
		Subscription: "some-kind-of-subscription-id",
	}

	params := apiService.ServiceAddParams{
		HTTPRequest: &http.Request{},
		ServiceItem: service,
	}

	checkResponse(srv.ServiceAddHandler(params, nil), t, 201, service)
	tc.handler.ValidateRequestCount(t, 2)
}

func TestServiceSubscriptionAlreadyExists(t *testing.T) {
	expectedObjects := &cloudlinuxv1alpha1.KuberLogicServiceList{
		Items: []cloudlinuxv1alpha1.KuberLogicService{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "advanced",
					Labels: map[string]string{
						"subscription-id": "existing-subscription-id",
					},
				},
				Spec: cloudlinuxv1alpha1.KuberLogicServiceSpec{
					Type: "postgresql",
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

	service := &models.Service{
		ID:           util.StrAsPointer("advanced"),
		Type:         util.StrAsPointer("postgresql"),
		Replicas:     util.Int64AsPointer(0),
		Status:       "Unknown",
		Subscription: "existing-subscription-id",
	}

	params := apiService.ServiceAddParams{
		HTTPRequest: &http.Request{},
		ServiceItem: service,
	}

	checkResponse(srv.ServiceAddHandler(params, nil), t, 400, &models.Error{
		Message: "Service with subscription 'existing-subscription-id' already exist",
	})
	tc.handler.ValidateRequestCount(t, 1)
}
