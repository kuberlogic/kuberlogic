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

func TestRestoreDelete(t *testing.T) {
	cases := []testCase{
		{
			name:   "ok",
			status: 200,
			objects: []runtime.Object{
				&v1alpha1.KuberlogicServiceRestore{
					ObjectMeta: v1.ObjectMeta{
						Name: "simple",
					},
					Spec: v1alpha1.KuberlogicServiceRestoreSpec{
						KuberlogicServiceBackup: "test",
					},
				},
			},
			result: &models.Restore{
				ID: "simple",
			},
			params: apiRestore.RestoreDeleteParams{
				HTTPRequest: &http.Request{},
				RestoreID:   "simple",
			},
		},
		{
			name:    "not-found",
			status:  404,
			objects: []runtime.Object{},
			result:  nil,
			params: apiRestore.RestoreDeleteParams{
				HTTPRequest: &http.Request{},
				RestoreID:   "simple",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkResponse(newFakeHandlers(t, tc.objects...).RestoreDeleteHandler(tc.params.(apiRestore.RestoreDeleteParams), nil), t, tc.status, tc.result)
		})
	}
}
