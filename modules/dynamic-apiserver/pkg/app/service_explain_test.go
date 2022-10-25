/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package app

import (
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"testing"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func TestServiceExplain(t *testing.T) {
	cases := []testCase{
		{
			name:   "service-not-found",
			status: 400,
			result: &models.Error{
				Message: "service does not exist",
			},
			params: apiService.ServiceExplainParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "service",
			},
		},
		{
			name:   "all-not-found",
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
			},
			result: &models.Explain{
				Pod: &models.ExplainPod{
					Error: "failed to get pod: no pod is available",
				},
				Pvc: &models.ExplainPvc{
					Error: "failed to get pvc: no pvc is available",
				},
				Ingress: &models.ExplainIngress{
					Error: "failed to get ingress: no ingress is available",
				},
			},
			params: apiService.ServiceExplainParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "logs-test",
			},
		},
		{
			name:   "ok",
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
						Namespace: "default",
					},
				},
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "default",
						Labels:    map[string]string{"docker-compose.service/name": "logs-test"},
					},
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							{
								Name:         "a",
								RestartCount: 5,
								State: v1.ContainerState{
									Running: &v1.ContainerStateRunning{},
								},
							},
							{
								Name:         "b",
								RestartCount: 0,
								State: v1.ContainerState{
									Waiting: &v1.ContainerStateWaiting{},
								},
							},
						},
					},
				},
				&v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pvc",
						Namespace: "default",
						Labels:    map[string]string{"docker-compose.service/name": "logs-test"},
					},
					Spec: v1.PersistentVolumeClaimSpec{
						StorageClassName: util.StrAsPointer("test"),
					},
					Status: v1.PersistentVolumeClaimStatus{
						Phase: v1.ClaimBound,
						Capacity: v1.ResourceList{
							v1.ResourceStorage: resource.MustParse("1Gi"),
						},
					},
				},
				&v12.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-ingress",
						Namespace: "default",
						Labels:    map[string]string{"docker-compose.service/name": "logs-test"},
					},
					Spec: v12.IngressSpec{
						IngressClassName: util.StrAsPointer("test-ingress-class"),
						Rules: []v12.IngressRule{
							{
								Host: "kuberlogic.com",
							}, {
								Host: "kuberlogic.org",
							},
						},
					},
				},
			},
			result: &models.Explain{
				Pod: &models.ExplainPod{
					Containers: []*models.ExplainPodContainersItems0{
						{
							Name:         "a",
							RestartCount: util.Int64AsPointer(5),
							Status:       "running",
						}, {
							Name:         "b",
							RestartCount: util.Int64AsPointer(0),
							Status:       "waiting",
						},
					},
				},
				Pvc: &models.ExplainPvc{
					StorageClass: "test",
					Phase:        "Bound",
					Size:         "1Gi",
				},
				Ingress: &models.ExplainIngress{
					Hosts:        []string{"kuberlogic.com", "kuberlogic.org"},
					IngressClass: "test-ingress-class",
				},
			},
			params: apiService.ServiceExplainParams{
				HTTPRequest: &http.Request{},
				ServiceID:   "logs-test",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkResponse(newFakeHandlers(t, tc.objects...).ServiceExplainHandler(tc.params.(apiService.ServiceExplainParams), nil), t, tc.status, tc.result)
		})
	}
}
