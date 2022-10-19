package app

import (
	"sort"

	"github.com/go-openapi/runtime/middleware"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
)

func (h *handlers) ServiceSecretsListHandler(params apiService.ServiceSecretsListParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	kls, err := h.Services().Get(ctx, params.ServiceID, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return apiService.NewServiceSecretsListBadRequest().WithPayload(&models.Error{
			Message: "service does not exist",
		})
	} else if err != nil {
		h.log.Errorw("failed to get service", "error", err.Error())
		return apiService.NewServiceSecretsListServiceUnavailable().WithPayload(&models.Error{
			Message: "failed to get service: " + err.Error(),
		})
	}

	// fixme: abstraction leak, secret is defined inside the plugin
	secretStorage, err := h.clientset.CoreV1().
		Secrets(kls.Status.Namespace).
		Get(params.HTTPRequest.Context(), kls.GetName(), metav1.GetOptions{})
	if err != nil {
		h.log.Errorw("failed to get service secret object", "error", err.Error())
		return apiService.NewServiceSecretsListServiceUnavailable().WithPayload(&models.Error{
			Message: "failed to retrieve service secrets",
		})
	}
	secrets := models.ServiceSecrets{}
	for secId, secData := range secretStorage.Data {
		secrets = append(secrets, &models.ServiceSecret{
			ID:    secId,
			Value: string(secData),
		})
	}
	sort.Slice(secrets, func(i, j int) bool {
		return secrets[i].ID < secrets[j].ID
	})

	return apiService.NewServiceSecretsListOK().WithPayload(secrets)
}
