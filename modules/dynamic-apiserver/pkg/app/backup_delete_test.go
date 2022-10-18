package app

import (
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	"net/http"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	apiBackup "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/backup"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBackupDelete(t *testing.T) {
	cases := []testCase{
		{
			name:   "ok",
			status: 200,
			objects: []runtime.Object{
				&v1alpha1.KuberlogicServiceBackup{
					ObjectMeta: v1.ObjectMeta{
						Name: "simple",
					},
					Spec: v1alpha1.KuberlogicServiceBackupSpec{
						KuberlogicServiceName: "test",
					},
				},
			},
			result: nil,
			params: apiBackup.BackupDeleteParams{
				HTTPRequest: &http.Request{},
				BackupID:    "simple",
			},
		},
		{
			name:    "not-found",
			status:  404,
			objects: []runtime.Object{},
			result: &models.Error{
				Message: "backup not found: simple",
			},
			params: apiBackup.BackupDeleteParams{
				HTTPRequest: &http.Request{},
				BackupID:    "simple",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkResponse(newFakeHandlers(t, tc.objects...).BackupDeleteHandler(tc.params.(apiBackup.BackupDeleteParams), nil), t, tc.status, tc.result)
		})
	}
}
