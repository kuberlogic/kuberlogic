package app

import (
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiBackup "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/backup"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	cloudlinuxv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"net/http"
	"testing"
	"time"
)

func TestBackupListEmpty(t *testing.T) {
	expectedObjects := &cloudlinuxv1alpha1.KuberlogicServiceBackupList{
		Items: []cloudlinuxv1alpha1.KuberlogicServiceBackup{},
	}

	tc := createTestClient(expectedObjects, 200, t)
	defer tc.server.Close()

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
	}

	params := apiBackup.BackupListParams{
		HTTPRequest: &http.Request{},
	}

	checkResponse(srv.BackupListHandler(params, nil), t, 200, models.Backups{})
	tc.handler.ValidateRequestCount(t, 1)
}

func TestBackupListMany(t *testing.T) {
	expectedObjects := &cloudlinuxv1alpha1.KuberlogicServiceBackupList{
		Items: []cloudlinuxv1alpha1.KuberlogicServiceBackup{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-%d", "service1", time.Now().Unix()),
				},
				Spec: cloudlinuxv1alpha1.KuberlogicServiceBackupSpec{
					KuberlogicServiceName: "service1",
				},
				Status: cloudlinuxv1alpha1.KuberlogicServiceBackupStatus{
					Phase: "Failed",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-%d", "service2", time.Now().Unix()),
				},
				Spec: cloudlinuxv1alpha1.KuberlogicServiceBackupSpec{
					KuberlogicServiceName: "service2",
				},
				Status: cloudlinuxv1alpha1.KuberlogicServiceBackupStatus{
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

	backups := models.Backups{
		{
			ID:        fmt.Sprintf("%s-%d", "service1", time.Now().Unix()),
			ServiceID: "service1",
			Status:    "Failed",
		},
		{
			ID:        fmt.Sprintf("%s-%d", "service2", time.Now().Unix()),
			ServiceID: "service2",
			Status:    "Successful",
		},
	}

	params := apiBackup.BackupListParams{
		HTTPRequest: &http.Request{},
	}

	checkResponse(srv.BackupListHandler(params, nil), t, 200, backups)
	tc.handler.ValidateRequestCount(t, 1)
}

func TestBackupListWithServiceFilter(t *testing.T) {
	expectedObjects := &cloudlinuxv1alpha1.KuberlogicServiceBackupList{
		Items: []cloudlinuxv1alpha1.KuberlogicServiceBackup{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-%d", "service1", time.Now().Unix()),
					Labels: map[string]string{
						"kls-id": "service1",
					},
				},
				Spec: cloudlinuxv1alpha1.KuberlogicServiceBackupSpec{
					KuberlogicServiceName: "service1",
				},
				Status: cloudlinuxv1alpha1.KuberlogicServiceBackupStatus{
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

	backups := models.Backups{
		{
			ID:        fmt.Sprintf("%s-%d", "service1", time.Now().Unix()),
			ServiceID: "service1",
			Status:    "Pending",
		},
	}

	params := apiBackup.BackupListParams{
		HTTPRequest: &http.Request{},
		ServiceID:   util.StrAsPointer("service1"),
	}

	checkResponse(srv.BackupListHandler(params, nil), t, 200, backups)
	tc.handler.ValidateRequestCount(t, 1)
}
