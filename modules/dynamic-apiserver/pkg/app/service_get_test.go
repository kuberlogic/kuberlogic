package app

import (
	"net/http"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	v11 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestServiceGet(t *testing.T) {
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
						Limits: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("2Gi"),
						},
					},
					Status: v1alpha1.KuberLogicServiceStatus{
						Phase: "Unknown",
					},
				},
			},
			result: &models.Service{
				ID:       util.StrAsPointer("one"),
				Type:     util.StrAsPointer("postgresql"),
				Replicas: util.Int64AsPointer(1),
				Limits: &models.Limits{
					Storage: "2Gi",
				},
				Status: "Unknown",
			},
			params: apiService.ServiceGetParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "one",
			},
		}, {
			name:   "not-found",
			status: 404,
			result: &models.Error{
				Message: "kuberlogic service not found: one",
			},
			params: apiService.ServiceGetParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "one",
			},
		}, {
			name:   "converting-error",
			status: 503,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicService{
					ObjectMeta: v1.ObjectMeta{
						Name: "one",
					},
					Spec: v1alpha1.KuberLogicServiceSpec{
						Type:     "postgresql",
						Replicas: 1,
						Limits: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("2Gi"),
						},
						Advanced: v11.JSON{Raw: []byte(`{"some": "invalid-json")}`)},
					},
					Status: v1alpha1.KuberLogicServiceStatus{
						Phase: "Unknown",
					},
				},
			},
			result: &models.Error{
				Message: "error converting kuberlogicservice: invalid character ')' after object key:value pair",
			},
			params: apiService.ServiceGetParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "one",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkResponse(newFakeHandlers(t, tc.objects...).ServiceGetHandler(tc.params.(apiService.ServiceGetParams), nil), t, tc.status, tc.result)
		})
	}
}
