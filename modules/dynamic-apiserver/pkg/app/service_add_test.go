package app

import (
	"net/http"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func TestServiceAdd(t *testing.T) {
	cases := []testCase{
		{
			name:    "ok-minimal",
			status:  201,
			objects: []runtime.Object{},
			result: &models.Service{
				ID:       util.StrAsPointer("simple"),
				Replicas: util.Int64AsPointer(1),
				Type:     util.StrAsPointer("postgresql"),
				Domain:   "simple.kuberlogic.local",
			},
			params: apiService.ServiceAddParams{
				HTTPRequest: &http.Request{},
				ServiceItem: &models.Service{
					ID:       util.StrAsPointer("simple"),
					Replicas: util.Int64AsPointer(1),
					Type:     util.StrAsPointer("postgresql"),
				},
			},
		},
		{
			name:    "ok-all-fields",
			status:  201,
			objects: []runtime.Object{},
			result: &models.Service{
				ID:             util.StrAsPointer("simple"),
				Replicas:       util.Int64AsPointer(1),
				Version:        "13",
				Type:           util.StrAsPointer("postgresql"),
				Domain:         "my-custom-domain.com",
				BackupSchedule: "7 * * * *",
				Limits: &models.Limits{
					CPU:     "250m",
					Memory:  "128Mi",
					Storage: "2Gi",
				},
				Insecure: true,
				Advanced: models.Advanced{
					"one": "1",
					"two": float64(2),
					"free": map[string]interface{}{
						"bool": true,
					},
				},
				Subscription: "some-kind-of-subscription-id",
			},
			params: apiService.ServiceAddParams{
				HTTPRequest: &http.Request{},
				ServiceItem: &models.Service{
					ID:             util.StrAsPointer("simple"),
					Replicas:       util.Int64AsPointer(1),
					Type:           util.StrAsPointer("postgresql"),
					Version:        "13",
					BackupSchedule: "7 * * * *",
					Domain:         "my-custom-domain.com",
					Limits: &models.Limits{
						CPU:     "250m",
						Memory:  "128Mi",
						Storage: "2Gi",
					},
					Insecure: true,
					Advanced: models.Advanced{
						"one": "1",
						"two": float64(2),
						"free": map[string]interface{}{
							"bool": true,
						},
					},
					Subscription: "some-kind-of-subscription-id",
				},
			},
		},
		{
			name:   "subscription-already-exists",
			status: 400,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicService{
					ObjectMeta: v1.ObjectMeta{
						Name: "existing",
						Labels: map[string]string{
							"subscription-id": "already-exists",
						},
					},
					Spec: v1alpha1.KuberLogicServiceSpec{
						Type: "postgresql",
					},
				},
			},
			result: &models.Error{
				Message: "service with subscription 'already-exists' already exist",
			},
			params: apiService.ServiceAddParams{
				HTTPRequest: &http.Request{},
				ServiceItem: &models.Service{
					ID:           util.StrAsPointer("new-service"),
					Replicas:     util.Int64AsPointer(1),
					Type:         util.StrAsPointer("postgresql"),
					Subscription: "already-exists",
				},
			},
		}, {
			name:   "to-model-converting-error",
			status: 400,
			result: &models.Error{
				Message: "cannot deserialize advanced parameter: json: unsupported type: func()",
			},
			params: apiService.ServiceAddParams{
				HTTPRequest: &http.Request{},
				ServiceItem: &models.Service{
					ID:           util.StrAsPointer("simple"),
					Replicas:     util.Int64AsPointer(1),
					Type:         util.StrAsPointer("postgresql"),
					Subscription: "already-exists",
					Advanced: models.Advanced{
						"key": func() {},
					},
				},
			},
		}, {
			name:   "already-exists",
			status: 409,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicService{
					ObjectMeta: v1.ObjectMeta{
						Name: "simple",
					},
					Spec: v1alpha1.KuberLogicServiceSpec{
						Type: "postgresql",
					},
				},
			},
			result: &models.Error{
				Message: "kuberlogic service already exists: simple",
			},
			params: apiService.ServiceAddParams{
				HTTPRequest: &http.Request{},
				ServiceItem: &models.Service{
					ID:       util.StrAsPointer("simple"),
					Replicas: util.Int64AsPointer(1),
					Type:     util.StrAsPointer("postgresql"),
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkResponse(newFakeHandlers(t, tc.objects...).ServiceAddHandler(tc.params.(apiService.ServiceAddParams), nil), t, tc.status, tc.result)
		})
	}
}
