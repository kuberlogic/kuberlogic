package app

import (
	"net/http"
	"testing"

	v11 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func TestServiceList(t *testing.T) {
	cases := []testCase{
		{
			name:   "one-service",
			status: 200,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicServiceList{
					Items: []v1alpha1.KuberLogicService{
						{
							ObjectMeta: v1.ObjectMeta{
								Name: "one",
								Labels: map[string]string{
									"subscription-id": "some-kind-of-subscription-id",
								},
							},
							Spec: v1alpha1.KuberLogicServiceSpec{
								Type:     "postgresql",
								Replicas: 1,
								Domain:   "example.com",
							},
							Status: v1alpha1.KuberLogicServiceStatus{
								Phase: "Running",
							},
						},
					},
				},
			},
			result: models.Services{
				{
					ID:           util.StrAsPointer("one"),
					Type:         util.StrAsPointer("postgresql"),
					Replicas:     util.Int64AsPointer(1),
					Status:       "Running",
					Subscription: "some-kind-of-subscription-id",
					Domain:       "example.com",
				},
			},
			params: apiService.ServiceListParams{HTTPRequest: &http.Request{}},
		}, {
			name:   "no-services",
			status: 200,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicServiceList{
					Items: []v1alpha1.KuberLogicService{},
				},
			},
			result: models.Services{},
			params: apiService.ServiceListParams{HTTPRequest: &http.Request{}},
		}, {
			name:   "many-services",
			status: 200,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicServiceList{
					Items: []v1alpha1.KuberLogicService{
						{
							ObjectMeta: v1.ObjectMeta{
								Name: "one",
							},
							Spec: v1alpha1.KuberLogicServiceSpec{
								Type:     "postgresql",
								Replicas: 1,
								Domain:   "example.com",
							},
							Status: v1alpha1.KuberLogicServiceStatus{
								Phase: "Running",
							},
						},
						{
							ObjectMeta: v1.ObjectMeta{
								Name: "two",
							},
							Spec: v1alpha1.KuberLogicServiceSpec{
								Type:     "mysql",
								Domain:   "example.com",
								Replicas: 2,
							},
							Status: v1alpha1.KuberLogicServiceStatus{
								Phase: "Failed",
							},
						},
					},
				},
			},
			result: models.Services{
				{
					ID:       util.StrAsPointer("one"),
					Type:     util.StrAsPointer("postgresql"),
					Replicas: util.Int64AsPointer(1),
					Status:   "Running",
					Domain:   "example.com",
				},
				{
					ID:       util.StrAsPointer("two"),
					Type:     util.StrAsPointer("mysql"),
					Replicas: util.Int64AsPointer(2),
					Status:   "Failed",
					Domain:   "example.com",
				},
			},
			params: apiService.ServiceListParams{HTTPRequest: &http.Request{}},
		}, {
			name:   "with-subscription-id",
			status: 200,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicServiceList{
					Items: []v1alpha1.KuberLogicService{
						{
							ObjectMeta: v1.ObjectMeta{
								Name: "one",
								Labels: map[string]string{
									"subscription-id": "some-kind-of-subscription-id",
								},
							},
							Spec: v1alpha1.KuberLogicServiceSpec{
								Type:     "postgresql",
								Replicas: 1,
								Domain:   "example.com",
							},
							Status: v1alpha1.KuberLogicServiceStatus{
								Phase: "Running",
							},
						},
						{
							ObjectMeta: v1.ObjectMeta{
								Name: "two",
								Labels: map[string]string{
									"subscription-id": "some-other-kind-of-subscription-id",
								},
							},
							Spec: v1alpha1.KuberLogicServiceSpec{
								Type:     "mysql",
								Replicas: 2,
								Domain:   "kuberlogic.com",
							},
							Status: v1alpha1.KuberLogicServiceStatus{
								Phase: "Failing",
							},
						},
					},
				},
			},
			result: models.Services{
				{
					ID:           util.StrAsPointer("one"),
					Type:         util.StrAsPointer("postgresql"),
					Replicas:     util.Int64AsPointer(1),
					Status:       "Running",
					Subscription: "some-kind-of-subscription-id",
					Domain:       "example.com",
				},
			},
			params: apiService.ServiceListParams{HTTPRequest: &http.Request{}, SubscriptionID: util.StrAsPointer("some-kind-of-subscription-id")},
		}, {
			name:   "convert-error",
			status: 503,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicServiceList{
					Items: []v1alpha1.KuberLogicService{
						{
							ObjectMeta: v1.ObjectMeta{
								Name: "one",
								Labels: map[string]string{
									"subscription-id": "some-kind-of-subscription-id",
								},
							},
							Spec: v1alpha1.KuberLogicServiceSpec{
								Type:     "postgresql",
								Replicas: 1,
								Domain:   "example.com",
								Advanced: v11.JSON{Raw: []byte(`{"some": "invalid-json")}`)},
							},
							Status: v1alpha1.KuberLogicServiceStatus{
								Phase: "Running",
							},
						},
					},
				},
			},
			result: &models.Error{
				Message: "error converting service object",
			},
			params: apiService.ServiceListParams{HTTPRequest: &http.Request{}},
		}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkResponse(
				newFakeHandlers(t, tc.objects...).
					ServiceListHandler(tc.params.(apiService.ServiceListParams), nil),
				t,
				tc.status,
				tc.result,
			)
		})
	}
}
