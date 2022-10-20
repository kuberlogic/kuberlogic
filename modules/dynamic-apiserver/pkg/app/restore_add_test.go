package app

import (
	"net/http"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiRestore "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/restore"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func TestRestoreAdd(t *testing.T) {
	cases := []testCase{
		{
			name:   "ok",
			status: 201,
			objects: []runtime.Object{
				&v1alpha1.KuberlogicServiceBackup{
					ObjectMeta: v1.ObjectMeta{
						Name: "existing-backup",
					},
					Spec: v1alpha1.KuberlogicServiceBackupSpec{
						KuberlogicServiceName: "test",
					},
				},
			},
			result: &models.Restore{
				BackupID: "existing-backup",
				ID:       "existing-backup",
			},
			params: apiRestore.RestoreAddParams{
				HTTPRequest: &http.Request{},
				RestoreItem: &models.Restore{
					BackupID: "existing-backup",
				},
			},
		},
		{
			name:   "already-exists",
			status: 409,
			objects: []runtime.Object{
				&v1alpha1.KuberlogicServiceBackup{
					ObjectMeta: v1.ObjectMeta{
						Name: "existing-backup",
					},
					Spec: v1alpha1.KuberlogicServiceBackupSpec{
						KuberlogicServiceName: "test",
					},
				},
				&v1alpha1.KuberlogicServiceRestore{
					ObjectMeta: v1.ObjectMeta{
						Name: "existing-backup",
					},
					Spec: v1alpha1.KuberlogicServiceRestoreSpec{
						KuberlogicServiceBackup: "existing-backup",
					},
				},
			},
			result: nil,
			params: apiRestore.RestoreAddParams{
				HTTPRequest: &http.Request{},
				RestoreItem: &models.Restore{
					BackupID: "existing-backup",
				},
			},
		},
		{
			name:    "backup-not-found",
			status:  400,
			objects: []runtime.Object{},
			result: &models.Error{
				Message: "backup `existing-backup` not found",
			},
			params: apiRestore.RestoreAddParams{
				HTTPRequest: &http.Request{},
				RestoreItem: &models.Restore{
					BackupID: "existing-backup",
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkResponse(newFakeHandlers(t, tc.objects...).RestoreAddHandler(tc.params.(apiRestore.RestoreAddParams), nil), t, tc.status, tc.result)
		})
	}
}
