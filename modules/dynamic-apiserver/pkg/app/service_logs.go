package app

import (
	"bytes"
	"io"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
)

func (h *handlers) ServiceLogsHandler(params apiService.ServiceLogsParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	kls, err := h.Services().Get(ctx, params.ServiceID, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return apiService.NewServiceLogsBadRequest().WithPayload(&models.Error{
			Message: "service does not exist",
		})
	} else if err != nil {
		e := errors.Wrap(err, "failed to get service")
		h.log.Errorw(e.Error())
		return apiService.NewServiceLogsServiceUnavailable().WithPayload(&models.Error{
			Message: e.Error(),
		})
	}

	listOptions := h.ListOptionsByKeyValue("docker-compose.service/name", &params.ServiceID)
	pods, err := h.clientset.CoreV1().Pods(kls.Status.Namespace).List(ctx, listOptions)
	if err != nil {
		e := errors.Wrap(err, "error getting service pods")
		h.log.Errorw(e.Error())
		return apiService.NewServiceLogsServiceUnavailable().WithPayload(&models.Error{
			Message: e.Error(),
		})
	}
	if len(pods.Items) < 1 {
		msg := "error listing service pods, no pod is available"
		h.log.Errorw(msg)
		return apiService.NewServiceLogsServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}
	pod := pods.Items[0] // filtering by labels should return only one pod
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
			e := errors.Wrapf(err, "error listing container '%s' logs", container.Name)
			h.log.Errorw(e.Error())
			return apiService.NewServiceLogsServiceUnavailable().WithPayload(&models.Error{
				Message: e.Error(),
			})
		}
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)

		if podLogsErr := podLogs.Close(); err != nil {
			e := errors.Wrapf(podLogsErr, "error closing pod logs stream for container '%s'", container.Name)
			h.log.Errorw(e.Error())
			return apiService.NewServiceLogsServiceUnavailable().WithPayload(&models.Error{
				Message: e.Error(),
			})
		}

		if err != nil {
			e := errors.Wrapf(err, "error obtaining container '%s' logs", container.Name)
			h.log.Errorw(e.Error())
			return apiService.NewServiceLogsServiceUnavailable().WithPayload(&models.Error{
				Message: e.Error(),
			})
		}
		response = append(response, &models.Log{ContainerName: container.Name, Logs: buf.String()})
	}
	return apiService.NewServiceLogsOK().WithPayload(response)
}
