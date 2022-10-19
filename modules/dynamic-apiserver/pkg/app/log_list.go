package app

import (
	"bytes"
	"fmt"
	"io"

	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiLogs "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/logs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	kuberlogicNamespace = "kuberlogic"
)

func (h *handlers) LogListHandler(params apiLogs.LogListParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()
	pods, err := h.clientset.CoreV1().Pods(kuberlogicNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		msg := "error listing kuberlogic pods"
		h.log.Errorw(msg)
		return apiLogs.NewLogListServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}
	if len(pods.Items) < 1 {
		msg := "error listing kuberlogic pods, no kuberlogic pod is available"
		h.log.Errorw(msg)
		return apiLogs.NewLogListServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}
	pod := pods.Items[0]
	response := models.Logs{}
	for _, container := range pod.Spec.Containers {
		if params.ContainerName != nil && container.Name != *params.ContainerName {
			continue
		}
		req := h.clientset.CoreV1().Pods(kuberlogicNamespace).GetLogs(pod.Name, &corev1.PodLogOptions{
			Container: container.Name,
		})
		podLogs, err := req.Stream(ctx)
		if err != nil {
			msg := fmt.Sprintf("error listing container '%s' logs", container.Name)
			h.log.Errorw(msg)
			h.log.Errorw(err.Error())
			return apiLogs.NewLogListServiceUnavailable().WithPayload(&models.Error{
				Message: msg,
			})
		}
		defer podLogs.Close()
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			msg := fmt.Sprintf("error obtaining container '%s' logs", container.Name)
			h.log.Errorw(msg)
			return apiLogs.NewLogListServiceUnavailable().WithPayload(&models.Error{
				Message: msg,
			})
		}
		response = append(response, &models.Log{ContainerName: container.Name, Logs: buf.String()})
	}
	return apiLogs.NewLogListOK().WithPayload(response)
}
