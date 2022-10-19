package app

import (
	"net/http"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	cloudlinuxv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func TestServiceSecretsList(t *testing.T) {
	cases := []testCase{
		{
			name:   "service-not-found",
			status: 400,
			result: &models.Error{
				Message: "service does not exist",
			},
			params: apiService.ServiceSecretsListParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "service",
			},
		}, {
			name: "secret-not-found",
			objects: []runtime.Object{
				&cloudlinuxv1alpha1.KuberLogicService{
					ObjectMeta: metav1.ObjectMeta{
						Name: "service",
					},
					Spec: cloudlinuxv1alpha1.KuberLogicServiceSpec{
						Type: "demo",
					},
					Status: cloudlinuxv1alpha1.KuberLogicServiceStatus{
						Namespace: "secrets-test",
					},
				},
			},
			status: 503,
			result: &models.Error{
				Message: "failed to retrieve service secrets",
			},
			params: apiService.ServiceSecretsListParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "service",
			},
		}, {
			name:   "empty",
			status: 200,
			objects: []runtime.Object{
				&cloudlinuxv1alpha1.KuberLogicService{
					ObjectMeta: metav1.ObjectMeta{
						Name: "secrets-test",
					},
					Spec: cloudlinuxv1alpha1.KuberLogicServiceSpec{
						Type: "demo",
					},
					Status: cloudlinuxv1alpha1.KuberLogicServiceStatus{
						Namespace: "secrets-test",
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secrets-test",
						Namespace: "secrets-test",
					},
					Data: map[string][]byte{},
				},
			},
			result: models.ServiceSecrets{},
			params: apiService.ServiceSecretsListParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "secrets-test",
			},
		}, {
			name:   "many",
			status: 200,
			objects: []runtime.Object{
				&cloudlinuxv1alpha1.KuberLogicService{
					ObjectMeta: metav1.ObjectMeta{
						Name: "secrets-test",
					},
					Spec: cloudlinuxv1alpha1.KuberLogicServiceSpec{
						Type: "demo",
					},
					Status: cloudlinuxv1alpha1.KuberLogicServiceStatus{
						Namespace: "secrets-test",
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secrets-test",
						Namespace: "secrets-test",
					},
					Data: map[string][]byte{
						"a": []byte("b"),
						"c": []byte("d"),
					},
				},
			},
			result: models.ServiceSecrets{
				{ID: "a", Value: "b"},
				{ID: "c", Value: "d"},
			},
			params: apiService.ServiceSecretsListParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "secrets-test",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkResponse(newFakeHandlers(t, tc.objects...).ServiceSecretsListHandler(tc.params.(apiService.ServiceSecretsListParams), nil), t, tc.status, tc.result)
		})
	}
}
