package app

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func TestServiceCredentialsUpdate(t *testing.T) {
	cases := []testCase{
		{
			name:   "ok",
			status: 200,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicService{
					ObjectMeta: v1.ObjectMeta{
						Name: "demo",
					},
					Status: v1alpha1.KuberLogicServiceStatus{
						Namespace: "default",
					},
				},
			},
			result: nil,
			params: apiService.ServiceCredentialsUpdateParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "demo",
				ServiceCredentials: map[string]string{
					"token": "demo",
				},
			},
			helpers: []func(args ...interface{}) error{
				deleteSecret,
			},
		}, {
			name:   "failed",
			status: 503,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicService{
					ObjectMeta: v1.ObjectMeta{
						Name: "demo",
					},
					Status: v1alpha1.KuberLogicServiceStatus{
						Namespace: "default",
					},
				},
			},
			result: &models.Error{
				Message: "failed to update credentials",
			},
			params: apiService.ServiceCredentialsUpdateParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "demo",
				ServiceCredentials: map[string]string{
					"token": "demo",
				},
			},
		}, {
			name:   "service-not-found",
			status: 400,
			result: &models.Error{
				Message: "service does not exist",
			},
			params: apiService.ServiceCredentialsUpdateParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "demo",
				ServiceCredentials: map[string]string{
					"token": "demo",
				},
			},
		},
		{
			name:   "credentials-already-exists",
			status: 503,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicService{
					ObjectMeta: v1.ObjectMeta{
						Name: "demo",
					},
					Status: v1alpha1.KuberLogicServiceStatus{
						Namespace: "default",
					},
				},
				&corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      v1alpha1.CredsUpdateSecretName,
						Namespace: "default",
					},
					StringData: map[string]string{
						"token": "demo",
					},
				},
			},
			result: &models.Error{
				Message: "failed to submit a credentials update request",
			},
			params: apiService.ServiceCredentialsUpdateParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "demo",
				ServiceCredentials: map[string]string{
					"token": "demo",
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			internalObjects, customObjects := splitObjects(tc.objects)
			clientset := fake.NewSimpleClientset(internalObjects...)
			h := newFakeHandlersWithClientset(t, clientset, customObjects...)

			wg := sync.WaitGroup{}
			wg.Add(2)
			go func() {
				defer wg.Done()
				for _, c := range tc.helpers {
					err := c(clientset)
					if err != nil {
						t.Error(err)
						return
					}
				}
			}()
			go func() {
				defer wg.Done()
				checkResponse(h.ServiceCredentialsUpdateHandler(tc.params.(apiService.ServiceCredentialsUpdateParams), nil), t, tc.status, tc.result)
			}()
			wg.Wait()
		})
	}
}

func deleteSecret(args ...interface{}) error {
	client := args[0].(kubernetes.Interface)
	time.Sleep(time.Second * 2) // wait for secret to be deleted
	return client.CoreV1().Secrets("default").Delete(context.TODO(), v1alpha1.CredsUpdateSecretName, v1.DeleteOptions{})
}
