package app

import (
	"net/http"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func TestServiceLogs(t *testing.T) {
	containerName := "a"

	cases := []testCase{
		{
			name:   "service-not-found",
			status: 400,
			result: &models.Error{
				Message: "service does not exist",
			},
			params: apiService.ServiceLogsParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "service",
			},
		},
		{
			name:   "pod-not-found",
			status: 503,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicService{
					ObjectMeta: metav1.ObjectMeta{
						Name: "logs-test",
					},
					Spec: v1alpha1.KuberLogicServiceSpec{
						Type: "demo",
					},
					Status: v1alpha1.KuberLogicServiceStatus{
						Namespace: "secrets-test",
					},
				},
			},
			result: &models.Error{
				Message: "error listing service pods, no pod is available",
			},
			params: apiService.ServiceLogsParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "logs-test",
			},
		},
		{
			name:   "all-containers",
			status: 200,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicService{
					ObjectMeta: metav1.ObjectMeta{
						Name: "logs-test",
					},
					Spec: v1alpha1.KuberLogicServiceSpec{
						Type: "demo",
					},
					Status: v1alpha1.KuberLogicServiceStatus{
						Namespace: "secrets-test",
					},
				},
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "secrets-test",
						Labels:    map[string]string{"docker-compose.service/name": "logs-test"},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{Name: "a"},
							{Name: "b"},
						},
					},
					Status: v1.PodStatus{},
				},
			},
			result: models.Logs{
				{ContainerName: "a", Logs: "fake logs"},
				{ContainerName: "b", Logs: "fake logs"},
			},
			params: apiService.ServiceLogsParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "logs-test",
			},
		},
		{
			name:   "single-container",
			status: 200,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicService{
					ObjectMeta: metav1.ObjectMeta{
						Name: "logs-test",
					},
					Spec: v1alpha1.KuberLogicServiceSpec{
						Type: "demo",
					},
					Status: v1alpha1.KuberLogicServiceStatus{
						Namespace: "secrets-test",
					},
				},
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "secrets-test",
						Labels:    map[string]string{"docker-compose.service/name": "logs-test"},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{Name: "a"},
							{Name: "b"},
						},
					},
					Status: v1.PodStatus{},
				},
			},
			result: models.Logs{
				{ContainerName: "a", Logs: "fake logs"},
			},
			params: apiService.ServiceLogsParams{
				HTTPRequest:   &http.Request{},
				ServiceID:     "logs-test",
				ContainerName: &containerName,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkResponse(newFakeHandlers(t, tc.objects...).ServiceLogsHandler(tc.params.(apiService.ServiceLogsParams), nil), t, tc.status, tc.result)
		})
	}
}
