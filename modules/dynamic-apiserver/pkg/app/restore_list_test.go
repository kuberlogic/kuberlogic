package app

import (
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiRestore "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/restore"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	cloudlinuxv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"net/http"
	"testing"
)

func TestRestoreListEmpty(t *testing.T) {
	expectedObjects := &cloudlinuxv1alpha1.KuberlogicServiceRestoreList{
		Items: []cloudlinuxv1alpha1.KuberlogicServiceRestore{},
	}

	tc := createTestClient(expectedObjects, 200, t)
	defer tc.server.Close()

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
	}

	params := apiRestore.RestoreListParams{
		HTTPRequest: &http.Request{},
	}

	checkResponse(srv.RestoreListHandler(params, nil), t, 200, models.Restores{})
	tc.handler.ValidateRequestCount(t, 1)
}

func TestRestoreListMany(t *testing.T) {
	expectedObjects := &cloudlinuxv1alpha1.KuberlogicServiceRestoreList{
		Items: []cloudlinuxv1alpha1.KuberlogicServiceRestore{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backup1",
				},
				Spec: cloudlinuxv1alpha1.KuberlogicServiceRestoreSpec{
					KuberlogicServiceBackup: "backup1",
				},
				Status: cloudlinuxv1alpha1.KuberlogicServiceRestoreStatus{
					Phase: "Failed",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backup2",
				},
				Spec: cloudlinuxv1alpha1.KuberlogicServiceRestoreSpec{
					KuberlogicServiceBackup: "backup2",
				},
				Status: cloudlinuxv1alpha1.KuberlogicServiceRestoreStatus{
					Phase: "Successful",
				},
			},
		},
	}

	tc := createTestClient(expectedObjects, 200, t)
	defer tc.server.Close()

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
	}

	backups := models.Restores{
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
	}

	params := apiRestore.RestoreListParams{
		HTTPRequest: &http.Request{},
	}

	checkResponse(srv.RestoreListHandler(params, nil), t, 200, backups)
	tc.handler.ValidateRequestCount(t, 1)
}

func TestRestoreListWithServiceFilter(t *testing.T) {
	expectedObjects := &cloudlinuxv1alpha1.KuberlogicServiceRestoreList{
		Items: []cloudlinuxv1alpha1.KuberlogicServiceRestore{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backup1",
					Labels: map[string]string{
						"kls-id": "service1",
					},
				},
				Spec: cloudlinuxv1alpha1.KuberlogicServiceRestoreSpec{
					KuberlogicServiceBackup: "backup1",
				},
				Status: cloudlinuxv1alpha1.KuberlogicServiceRestoreStatus{
					Phase: "Pending",
				},
			},
		},
	}

	tc := createTestClient(expectedObjects, 200, t)
	defer tc.server.Close()

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
	}

	backups := models.Restores{
		{
			ID:       "backup1",
			BackupID: "backup1",
			Status:   "Pending",
		},
	}

	params := apiRestore.RestoreListParams{
		HTTPRequest: &http.Request{},
		ServiceID:   util.StrAsPointer("service1"),
	}

	checkResponse(srv.RestoreListHandler(params, nil), t, 200, backups)
	tc.handler.ValidateRequestCount(t, 1)
}
