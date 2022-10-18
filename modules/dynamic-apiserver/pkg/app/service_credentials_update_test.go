package app

import (
	"context"
	"net/http"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func TestServiceCredentialsUpdateOK(t *testing.T) {
	tc := createTestClient(&v1alpha1.KuberLogicService{
		ObjectMeta: v1.ObjectMeta{
			Name: "demo",
		},
		Status: v1alpha1.KuberLogicServiceStatus{
			Namespace: "default",
		},
	}, 200, t)
	defer tc.server.Close()

	fakeclientset := fake.NewSimpleClientset()
	srv := New(nil, fakeclientset, tc.client, &TestLog{t: t})

	params := apiService.ServiceCredentialsUpdateParams{
		HTTPRequest: &http.Request{},
		ServiceID:   "service",
		ServiceCredentials: map[string]string{
			"token": "demo",
		},
	}

	// simulate cred request fulfillment
	go func() {
		time.Sleep(time.Second * 2)
		_ = fakeclientset.CoreV1().Secrets("default").Delete(context.TODO(), v1alpha1.CredsUpdateSecretName, v1.DeleteOptions{})
	}()
	checkResponse(srv.ServiceCredentialsUpdateHandler(params, nil), t, 200, nil)
}

func TestServiceCredentialsUpdateFailed(t *testing.T) {
	tc := createTestClient(&v1alpha1.KuberLogicService{
		ObjectMeta: v1.ObjectMeta{
			Name: "demo",
		},
		Status: v1alpha1.KuberLogicServiceStatus{
			Namespace: "default",
		},
	}, 200, t)
	defer tc.server.Close()

	fakeclientset := fake.NewSimpleClientset()
	srv := New(nil, fakeclientset, tc.client, &TestLog{t: t})

	params := apiService.ServiceCredentialsUpdateParams{
		HTTPRequest: &http.Request{},
		ServiceID:   "service",
		ServiceCredentials: map[string]string{
			"token": "demo",
		},
	}

	checkResponse(srv.ServiceCredentialsUpdateHandler(params, nil), t, 503, &models.Error{
		Message: "failed to update credentials",
	})
}
