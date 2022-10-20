package app

import (
	"bytes"
	"fmt"
	"io"

	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h *handlers) ServiceLogsListHandler(params apiService.ServiceLogsListParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	kls, err := h.Services().Get(ctx, params.ServiceID, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return apiService.NewServiceSecretsListBadRequest().WithPayload(&models.Error{
			Message: "service does not exist",
		})
	} else if err != nil {
		h.log.Errorw("failed to get service", "error", err.Error())
		return apiService.NewServiceLogsListServiceUnavailable().WithPayload(&models.Error{
			Message: "failed to get service: " + err.Error(),
		})
	}

	pods, err := h.clientset.CoreV1().Pods(kls.Status.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		msg := "error listing kuberlogic service pods"
		h.log.Errorw(msg)
		return apiService.NewServiceLogsListServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}
	if len(pods.Items) < 1 {
		msg := "error listing service pods, no pod is available"
		h.log.Errorw(msg)
		return apiService.NewServiceLogsListServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}
	pod := pods.Items[0]
	response := models.Logs{}
	for _, container := range pod.Spec.Containers {
		if params.ContainerName != nil && container.Name != *params.ContainerName {
			continue
		}
		req := h.clientset.CoreV1().Pods(kls.Status.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
			Container: container.Name,
		})
		podLogs, err := req.Stream(ctx)
		if err != nil {
			msg := fmt.Sprintf("error listing container '%s' logs", container.Name)
			h.log.Errorw(msg)
			h.log.Errorw(err.Error())
			return apiService.NewServiceLogsListServiceUnavailable().WithPayload(&models.Error{
				Message: msg,
			})
		}
		defer podLogs.Close()
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			msg := fmt.Sprintf("error obtaining container '%s' logs", container.Name)
			h.log.Errorw(msg)
			return apiService.NewServiceLogsListServiceUnavailable().WithPayload(&models.Error{
				Message: msg,
			})
		}
		response = append(response, &models.Log{ContainerName: container.Name, Logs: buf.String()})
	}
	return apiService.NewServiceLogsListOK().WithPayload(response)
}
