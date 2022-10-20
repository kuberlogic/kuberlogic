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

func TestServiceEdit(t *testing.T) {
	cases := []testCase{
		{
			name:   "ok",
			status: 200,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicService{
					ObjectMeta: v1.ObjectMeta{
						Name: "one",
					},
					Spec: v1alpha1.KuberLogicServiceSpec{
						Type:     "postgresql",
						Replicas: 1,
					},
				},
			},
			result: nil,
			params: apiService.ServiceEditParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "one",
				ServiceItem: &models.Service{
					ID:       util.StrAsPointer("one"),
					Type:     util.StrAsPointer("postgresql"),
					Replicas: util.Int64AsPointer(2),
				},
			},
		},
		{
			name:   "subscription-can-not-be-changed",
			status: 400,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicService{
					ObjectMeta: v1.ObjectMeta{
						Name: "one",
					},
					Spec: v1alpha1.KuberLogicServiceSpec{
						Type:     "postgresql",
						Replicas: 1,
					},
				},
			},
			result: &models.Error{
				Message: "subscription cannot be changed",
			},
			params: apiService.ServiceEditParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "one",
				ServiceItem: &models.Service{
					ID:           util.StrAsPointer("one"),
					Type:         util.StrAsPointer("postgresql"),
					Replicas:     util.Int64AsPointer(2),
					Subscription: "new-subscription",
				},
			},
		}, {
			name:   "converting-error",
			status: 400,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicService{
					ObjectMeta: v1.ObjectMeta{
						Name: "broken-advanced-field",
					},
					Spec: v1alpha1.KuberLogicServiceSpec{
						Type:     "postgresql",
						Replicas: 1,
					},
				},
			},
			result: &models.Error{
				Message: "cannot deserialize advanced parameter: json: unsupported type: func()",
			},
			params: apiService.ServiceEditParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "broken-advanced-field",
				ServiceItem: &models.Service{
					ID:       util.StrAsPointer("broken-advanced-field"),
					Type:     util.StrAsPointer("postgresql"),
					Replicas: util.Int64AsPointer(2),
					Advanced: models.Advanced{
						"key": func() {},
					},
				},
			},
		}, {
			name:   "not-found",
			status: 404,
			result: &models.Error{
				Message: "kuberlogic service not found: not-found-id",
			},
			params: apiService.ServiceEditParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "not-found-id",
				ServiceItem: &models.Service{
					ID:       util.StrAsPointer("not-found-id"),
					Type:     util.StrAsPointer("postgresql"),
					Replicas: util.Int64AsPointer(2),
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkResponse(newFakeHandlers(t, tc.objects...).ServiceEditHandler(tc.params.(apiService.ServiceEditParams), nil), t, tc.status, tc.result)
		})
	}
}
