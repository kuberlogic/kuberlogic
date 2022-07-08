package app

import (
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiBackup "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/backup"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"net/http"
	"testing"
	"time"
)

func TestBackupAdd(t *testing.T) {
	expectedKlb := &kuberlogiccomv1alpha1.KuberlogicServiceBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%d", "test", time.Now().Unix()),
		},
		Spec: kuberlogiccomv1alpha1.KuberlogicServiceBackupSpec{
			KuberlogicServiceName: "test",
		},
	}

	tc := createTestClient(expectedKlb, 200, t)

	srv := &Service{
		log:              &TestLog{t: t},
		clientset:        fake.NewSimpleClientset(),
		kuberlogicClient: tc.client,
	}

	backup := &models.Backup{
		ServiceID: "test",
	}

	params := apiBackup.BackupAddParams{
		HTTPRequest: &http.Request{},
		BackupItem:  backup,
	}

	expectedBackup := &models.Backup{
		ServiceID: "test",
		ID:        expectedKlb.GetName(),
	}

	checkResponse(srv.BackupAddHandler(params, nil), t, 201, expectedBackup)
	tc.handler.ValidateRequestCount(t, 2)
}
