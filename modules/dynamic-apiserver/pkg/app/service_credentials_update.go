package app

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func (h *handlers) ServiceCredentialsUpdateHandler(params apiService.ServiceCredentialsUpdateParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	// check for service existence
	kls, err := h.Services().Get(ctx, params.ServiceID, v1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return apiService.NewServiceCredentialsUpdateBadRequest().WithPayload(&models.Error{
			Message: "service does not exist",
		})
	} else if err != nil {
		h.log.Errorw("failed to get service", "error", err.Error())
		return apiService.NewServiceCredentialsUpdateServiceUnavailable()
	}

	// create a credential secret
	credentialsUpdateRequest := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      v1alpha1.CredsUpdateSecretName,
			Namespace: kls.Status.Namespace,
			OwnerReferences: []v1.OwnerReference{
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
	credentialsUpdateRequest, err = h.clientset.CoreV1().Secrets(credentialsUpdateRequest.GetNamespace()).
		Create(ctx, credentialsUpdateRequest, v1.CreateOptions{FieldManager: "kuberlogic"})
	if err != nil {
		h.log.Errorw("failed to create a credentials update request secret", "error", err.Error())
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
		if _, err := h.clientset.CoreV1().Secrets(credentialsUpdateRequest.GetNamespace()).
			Get(ctx, credentialsUpdateRequest.GetName(), v1.GetOptions{}); k8serrors.IsNotFound(err) {
			return apiService.NewServiceCredentialsUpdateOK()
		} else if err != nil {
			h.log.Errorw("failed to get credentials update request secret", "error", err.Error())
			return apiService.NewServiceCredentialsUpdateServiceUnavailable().WithPayload(&models.Error{
				Message: "error waiting for a credentials update request fulfillment",
			})
		}
	}

	if err := h.clientset.CoreV1().Secrets(credentialsUpdateRequest.GetNamespace()).
		Delete(ctx, credentialsUpdateRequest.GetName(), v1.DeleteOptions{}); err != nil {
		h.log.Errorw("failed to delete credentials update request secret", "error", err.Error())
	}
	return apiService.NewServiceCredentialsUpdateServiceUnavailable().WithPayload(&models.Error{
		Message: "failed to update credentials",
	})
}
