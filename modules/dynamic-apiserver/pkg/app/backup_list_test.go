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
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func TestBackupList(t *testing.T) {
	cases := []testCase{
		{
			name:   "empty",
			status: 200,
			objects: []runtime.Object{
				&v1alpha1.KuberlogicServiceBackupList{
					Items: []v1alpha1.KuberlogicServiceBackup{},
				},
			},
			result: models.Backups{},
			params: apiBackup.BackupListParams{
				HTTPRequest: &http.Request{},
				ServiceID:   util.StrAsPointer("test-service"),
			},
		},
		{
			name:   "many",
			status: 200,
			objects: []runtime.Object{
				&v1alpha1.KuberlogicServiceBackup{
					ObjectMeta: v1.ObjectMeta{
						Name: fmt.Sprintf("%s-%d", "service1", time.Now().Unix()),
						Labels: map[string]string{
							util.BackupRestoreServiceField: "target-service",
						},
					},
					Spec: v1alpha1.KuberlogicServiceBackupSpec{
						KuberlogicServiceName: "target-service",
					},
					Status: v1alpha1.KuberlogicServiceBackupStatus{
						Phase: "Failed",
					},
				},
				&v1alpha1.KuberlogicServiceBackup{
					ObjectMeta: v1.ObjectMeta{
						Name: fmt.Sprintf("%s-%d", "service2", time.Now().Unix()),
						Labels: map[string]string{
							util.BackupRestoreServiceField: "another-service",
						},
					},
					Spec: v1alpha1.KuberlogicServiceBackupSpec{
						KuberlogicServiceName: "another-service",
					},
					Status: v1alpha1.KuberlogicServiceBackupStatus{
						Phase: "Successful",
					},
				},
				&v1alpha1.KuberlogicServiceBackup{
					ObjectMeta: v1.ObjectMeta{
						Name: fmt.Sprintf("%s-%d", "service3", time.Now().Unix()),
						Labels: map[string]string{
							util.BackupRestoreServiceField: "target-service",
						},
					},
					Spec: v1alpha1.KuberlogicServiceBackupSpec{
						KuberlogicServiceName: "target-service",
					},
					Status: v1alpha1.KuberlogicServiceBackupStatus{
						Phase: "Successful",
					},
				},
			},
			result: models.Backups{
				{
					ID:        fmt.Sprintf("%s-%d", "service1", time.Now().Unix()),
					ServiceID: "target-service",
					Status:    "Failed",
				},
				{
					ID:        fmt.Sprintf("%s-%d", "service2", time.Now().Unix()),
					ServiceID: "another-service",
					Status:    "Successful",
				},
				{
					ID:        fmt.Sprintf("%s-%d", "service3", time.Now().Unix()),
					ServiceID: "target-service",
					Status:    "Successful",
				},
			},
			params: apiBackup.BackupListParams{
				HTTPRequest: &http.Request{},
			},
		},
		{
			name:   "filtered-many",
			status: 200,
			objects: []runtime.Object{
				&v1alpha1.KuberlogicServiceBackup{
					ObjectMeta: v1.ObjectMeta{
						Name: fmt.Sprintf("%s-%d", "service1", time.Now().Unix()),
						Labels: map[string]string{
							util.BackupRestoreServiceField: "target-service",
						},
					},
					Spec: v1alpha1.KuberlogicServiceBackupSpec{
						KuberlogicServiceName: "target-service",
					},
					Status: v1alpha1.KuberlogicServiceBackupStatus{
						Phase: "Failed",
					},
				},
				&v1alpha1.KuberlogicServiceBackup{
					ObjectMeta: v1.ObjectMeta{
						Name: fmt.Sprintf("%s-%d", "service2", time.Now().Unix()),
						Labels: map[string]string{
							util.BackupRestoreServiceField: "another-service",
						},
					},
					Spec: v1alpha1.KuberlogicServiceBackupSpec{
						KuberlogicServiceName: "another-service",
					},
					Status: v1alpha1.KuberlogicServiceBackupStatus{
						Phase: "Successful",
					},
				},
				&v1alpha1.KuberlogicServiceBackup{
					ObjectMeta: v1.ObjectMeta{
						Name: fmt.Sprintf("%s-%d", "service3", time.Now().Unix()),
						Labels: map[string]string{
							util.BackupRestoreServiceField: "target-service",
						},
					},
					Spec: v1alpha1.KuberlogicServiceBackupSpec{
						KuberlogicServiceName: "target-service",
					},
					Status: v1alpha1.KuberlogicServiceBackupStatus{
						Phase: "Successful",
					},
				},
			},
			result: models.Backups{
				{
					ID:        fmt.Sprintf("%s-%d", "service1", time.Now().Unix()),
					ServiceID: "target-service",
					Status:    "Failed",
				},
				{
					ID:        fmt.Sprintf("%s-%d", "service3", time.Now().Unix()),
					ServiceID: "target-service",
					Status:    "Successful",
				},
			},
			params: apiBackup.BackupListParams{
				HTTPRequest: &http.Request{},
				ServiceID:   util.StrAsPointer("target-service"),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkResponse(newFakeHandlers(t, tc.objects...).BackupListHandler(tc.params.(apiBackup.BackupListParams), nil), t, tc.status, tc.result)
		})
	}
}
