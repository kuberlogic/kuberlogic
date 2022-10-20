package app

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiBackup "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/backup"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func TestBackupAdd(t *testing.T) {
	serviceName := "existing-service"
	backupName := fmt.Sprintf("%s-%d", serviceName, time.Now().Unix())
	cases := []testCase{
		{
			name:   "ok",
			status: 201,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicService{
					ObjectMeta: v1.ObjectMeta{
						Name: serviceName,
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
			result: &models.Backup{
				ServiceID: serviceName,
				ID:        backupName,
			},
			params: apiBackup.BackupAddParams{
				HTTPRequest: &http.Request{},
				BackupItem: &models.Backup{
					ServiceID: serviceName,
				},
			},
		},
		{
			name:    "service-not-found",
			status:  400,
			objects: []runtime.Object{},
			result: &models.Error{
				Message: fmt.Sprintf("service `%s` not found", serviceName),
			},
			params: apiBackup.BackupAddParams{
				HTTPRequest: &http.Request{},
				BackupItem: &models.Backup{
					ServiceID: serviceName,
				},
			},
		},
		{
			name:   "already-exists",
			status: 409,
			objects: []runtime.Object{
				&v1alpha1.KuberlogicServiceBackup{
					ObjectMeta: v1.ObjectMeta{
						Name: backupName,
					},
					Spec: v1alpha1.KuberlogicServiceBackupSpec{
						KuberlogicServiceName: serviceName,
					},
				},
				&v1alpha1.KuberLogicService{
					ObjectMeta: v1.ObjectMeta{
						Name: serviceName,
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
			result: nil,
			params: apiBackup.BackupAddParams{
				HTTPRequest: &http.Request{},
				BackupItem: &models.Backup{
					ServiceID: serviceName,
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkResponse(newFakeHandlers(t, tc.objects...).BackupAddHandler(tc.params.(apiBackup.BackupAddParams), nil), t, tc.status, tc.result)
		})
	}
}
