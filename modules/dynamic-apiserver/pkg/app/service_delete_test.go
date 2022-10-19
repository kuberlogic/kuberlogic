package app

import (
	"net/http"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func TestServiceDelete(t *testing.T) {
	cases := []testCase{
		{
			name:   "ok",
			status: 200,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicService{
					ObjectMeta: v1.ObjectMeta{
						Name: "service",
					},
					Spec: v1alpha1.KuberLogicServiceSpec{
						Type:     "postgresql",
						Replicas: 1,
					},
				},
			},
			result: &models.Error{
				Message: "kuberlogic service not found: service",
			},
			params: apiService.ServiceDeleteParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "service",
			},
		}, {
			name:   "not-found",
			status: 404,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicService{
					ObjectMeta: v1.ObjectMeta{
						Name: "other-service",
					},
					Spec: v1alpha1.KuberLogicServiceSpec{
						Type:     "postgresql",
						Replicas: 1,
					},
				},
			},
			result: &models.Error{
				Message: "kuberlogic service not found: service",
			},
			params: apiService.ServiceDeleteParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "service",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkResponse(newFakeHandlers(t, tc.objects...).ServiceDeleteHandler(tc.params.(apiService.ServiceDeleteParams), nil), t, tc.status, tc.result)
		})
	}
}
