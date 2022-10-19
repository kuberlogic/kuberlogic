package app

import (
	"net/http"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/config"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	cloudlinuxv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func TestServiceSecretsListEmpty(t *testing.T) {
	expectedObject := &cloudlinuxv1alpha1.KuberLogicService{
		ObjectMeta: metav1.ObjectMeta{
			Name: "secrets-test",
		},
		Spec: cloudlinuxv1alpha1.KuberLogicServiceSpec{
			Type: "demo",
		},
		Status: cloudlinuxv1alpha1.KuberLogicServiceStatus{
			Namespace: "secrets-test",
		},
	}
	expectedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      expectedObject.GetName(),
			Namespace: expectedObject.Status.Namespace,
		},
		Data: map[string][]byte{},
	}

	tc := createTestClient(expectedObject, 200, t)
	defer tc.server.Close()

	srv := New(&config.Config{
		Domain: "example.com",
	}, fake.NewSimpleClientset(expectedSecret), tc.client, &TestLog{t: t})

	params := apiService.ServiceSecretsListParams{
		HTTPRequest: &http.Request{},
		ServiceID:   expectedObject.GetName(),
	}

	checkResponse(srv.ServiceSecretsListHandler(params, nil), t, 200, models.ServiceSecrets{})
	tc.handler.ValidateRequestCount(t, 1)
}

func TestServiceSecretsListMany(t *testing.T) {
	expectedObject := &cloudlinuxv1alpha1.KuberLogicService{
		ObjectMeta: metav1.ObjectMeta{
			Name: "secrets-test",
		},
		Spec: cloudlinuxv1alpha1.KuberLogicServiceSpec{
			Type: "demo",
		},
		Status: cloudlinuxv1alpha1.KuberLogicServiceStatus{
			Namespace: "secrets-test",
		},
	}
	expectedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      expectedObject.GetName(),
			Namespace: expectedObject.Status.Namespace,
		},
		Data: map[string][]byte{
			"a": []byte("b"),
			"c": []byte("d"),
		},
	}

	tc := createTestClient(expectedObject, 200, t)
	defer tc.server.Close()

	srv := New(&config.Config{
		Domain: "example.com",
	}, fake.NewSimpleClientset(expectedSecret), tc.client, &TestLog{t: t})

	params := apiService.ServiceSecretsListParams{
		HTTPRequest: &http.Request{},
		ServiceID:   expectedObject.GetName(),
	}

	secrets := models.ServiceSecrets{
		{
			ID:    "a",
			Value: "b",
		},
		{
			ID:    "c",
			Value: "d",
		},
	}
	checkResponse(srv.ServiceSecretsListHandler(params, nil), t, 200, secrets)
	tc.handler.ValidateRequestCount(t, 1)
}
