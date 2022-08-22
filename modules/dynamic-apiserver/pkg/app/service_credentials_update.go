package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"time"
)

func (srv *Service) ServiceCredentialsUpdateHandler(params apiService.ServiceCredentialsUpdateParams, _ *models.Principal) middleware.Responder {
	// check for service existence
	kls := new(v1alpha1.KuberLogicService)
	if err := srv.kuberlogicClient.Get().
		Name(params.ServiceID).
		Resource(serviceK8sResource).
		Do(params.HTTPRequest.Context()).
		Into(kls); k8serrors.IsNotFound(err) {
		return apiService.NewServiceCredentialsUpdateBadRequest().WithPayload(&models.Error{
			Message: "service does not exist",
		})
	} else if err != nil {
		srv.log.Errorw("failed to get service", "error", err.Error())
		return apiService.NewServiceCredentialsUpdateServiceUnavailable()
	}

	// create a credential secret
	credentialsUpdateRequest := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      v1alpha1.CredsUpdateSecretName,
			Namespace: kls.Status.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         kls.APIVersion,
					Kind:               kls.GetObjectKind().GroupVersionKind().Kind,
					Name:               kls.GetName(),
					UID:                kls.GetUID(),
					BlockOwnerDeletion: pointer.BoolPtr(true),
					Controller:         pointer.BoolPtr(true),
				},
			},
		},
		StringData: params.ServiceCredentials,
	}
	credentialsUpdateRequest, err := srv.clientset.CoreV1().Secrets(credentialsUpdateRequest.GetNamespace()).
		Create(params.HTTPRequest.Context(), credentialsUpdateRequest, metav1.CreateOptions{FieldManager: "kuberlogic"})
	if err != nil {
		srv.log.Errorw("failed to create a credentials update request secret", "error", err.Error())
		return apiService.NewServiceCredentialsUpdateServiceUnavailable().WithPayload(&models.Error{
			Message: "failed to submit a credentials update request",
		})
	}

	// deletion means success
	// a better way would be to watch for the secret deletion,
	// but it may end up in a race condition when a secret is already deleted before this step
	timeout := time.Second * 15
	waitstep := time.Millisecond * 100
	for i := time.Duration(0); i < timeout; i += waitstep {
		time.Sleep(waitstep)
		if _, err := srv.clientset.CoreV1().Secrets(credentialsUpdateRequest.GetNamespace()).
			Get(params.HTTPRequest.Context(), credentialsUpdateRequest.GetName(), metav1.GetOptions{}); k8serrors.IsNotFound(err) {
			return apiService.NewServiceCredentialsUpdateOK()
		} else if err != nil {
			srv.log.Errorw("failed to get credentials update request secret", "error", err.Error())
			return apiService.NewServiceCredentialsUpdateServiceUnavailable().WithPayload(&models.Error{
				Message: "error waiting for a credentials update request fulfillment",
			})
		}
	}

	if err := srv.clientset.CoreV1().Secrets(credentialsUpdateRequest.GetNamespace()).
		Delete(params.HTTPRequest.Context(), credentialsUpdateRequest.GetName(), metav1.DeleteOptions{}); err != nil {
		srv.log.Errorw("failed to delete credentials update request secret", "error", err.Error())
	}
	return apiService.NewServiceCredentialsUpdateServiceUnavailable().WithPayload(&models.Error{
		Message: "failed to update credentials",
	})
}
