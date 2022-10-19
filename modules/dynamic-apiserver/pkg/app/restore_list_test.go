package app

import (
	"net/http"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiRestore "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/restore"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func TestRestoreList(t *testing.T) {
	cases := []testCase{
		{
			name:    "empty",
			status:  200,
			objects: []runtime.Object{},
			result:  models.Restores{},
			params: apiRestore.RestoreListParams{
				HTTPRequest: &http.Request{},
			},
		},
		{
			name:   "many",
			status: 200,
			objects: []runtime.Object{
				&v1alpha1.KuberlogicServiceRestore{
					ObjectMeta: v1.ObjectMeta{
						Name: "backup1",
					},
					Spec: v1alpha1.KuberlogicServiceRestoreSpec{
						KuberlogicServiceBackup: "backup1",
					},
					Status: v1alpha1.KuberlogicServiceRestoreStatus{
						Phase: "Failed",
					},
				},
				&v1alpha1.KuberlogicServiceRestore{
					ObjectMeta: v1.ObjectMeta{
						Name: "backup2",
					},
					Spec: v1alpha1.KuberlogicServiceRestoreSpec{
						KuberlogicServiceBackup: "backup2",
					},
					Status: v1alpha1.KuberlogicServiceRestoreStatus{
						Phase: "Successful",
					},
				},
			},
			result: models.Restores{
				{
					ID:       "backup1",
					BackupID: "backup1",
					Status:   "Failed",
				},
				{
					ID:       "backup2",
					BackupID: "backup2",
					Status:   "Successful",
				},
			},
			params: apiRestore.RestoreListParams{
				HTTPRequest: &http.Request{},
			},
		},
		{
			name:   "filtered-many",
			status: 200,
			objects: []runtime.Object{
				&v1alpha1.KuberlogicServiceRestore{
					ObjectMeta: v1.ObjectMeta{
						Name: "backup1",
						Labels: map[string]string{
							util.BackupRestoreServiceField: "target-service",
						},
					},
					Spec: v1alpha1.KuberlogicServiceRestoreSpec{
						KuberlogicServiceBackup: "backup1",
					},
					Status: v1alpha1.KuberlogicServiceRestoreStatus{
						Phase: "Failed",
					},
				},
				&v1alpha1.KuberlogicServiceRestore{
					ObjectMeta: v1.ObjectMeta{
						Name: "backup2",
						Labels: map[string]string{
							util.BackupRestoreServiceField: "another-service",
						},
					},
					Spec: v1alpha1.KuberlogicServiceRestoreSpec{
						KuberlogicServiceBackup: "backup2",
					},
					Status: v1alpha1.KuberlogicServiceRestoreStatus{
						Phase: "Successful",
					},
				},
				&v1alpha1.KuberlogicServiceRestore{
					ObjectMeta: v1.ObjectMeta{
						Name: "backup3",
						Labels: map[string]string{
							util.BackupRestoreServiceField: "target-service",
						},
					},
					Spec: v1alpha1.KuberlogicServiceRestoreSpec{
						KuberlogicServiceBackup: "backup1",
					},
					Status: v1alpha1.KuberlogicServiceRestoreStatus{
						Phase: "Successful",
					},
				},
			},
			result: models.Restores{
				{
					ID:       "backup1",
					BackupID: "backup1",
					Status:   "Failed",
				},
				{
					ID:       "backup3",
					BackupID: "backup1",
					Status:   "Successful",
				},
			},
			params: apiRestore.RestoreListParams{
				HTTPRequest: &http.Request{},
				ServiceID:   util.StrAsPointer("target-service"),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkResponse(newFakeHandlers(t, tc.objects...).RestoreListHandler(tc.params.(apiRestore.RestoreListParams), nil), t, tc.status, tc.result)
		})
	}
}
