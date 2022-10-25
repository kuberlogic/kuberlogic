package app

import (
	"context"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"

	"github.com/go-openapi/runtime/middleware"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
)

func (h *handlers) ServiceExplainHandler(params apiService.ServiceExplainParams, _ *models.Principal) middleware.Responder {
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
	result := new(models.Explain)
	pod, err := firstPod(ctx, h.clientset, kls.Namespace, listOptions)
	if err != nil {
		result.Pod = &models.ExplainPod{
			Error: errors.Wrap(err, "failed to get pod").Error(),
		}
	} else {
		containersInfo := make([]*models.ExplainPodContainersItems0, 0)
		for _, c := range pod.Status.ContainerStatuses {
			containersInfo = append(containersInfo, &models.ExplainPodContainersItems0{
				Name:         c.Name,
				RestartCount: util.Int64AsPointer(int64(c.RestartCount)),
				Status:       containerStatus(c.State),
			})
		}

		result.Pod = &models.ExplainPod{
			Containers: containersInfo,
		}
	}

	pvc, err := firstPVC(ctx, h.clientset, kls.Namespace, listOptions)
	if err != nil {
		result.Pvc = &models.ExplainPvc{
			Error: errors.Wrap(err, "failed to get pvc").Error(),
		}
	} else {
		result.Pvc = &models.ExplainPvc{
			StorageClass: *pvc.Spec.StorageClassName,
			Phase:        string(pvc.Status.Phase),
			Size:         pvc.Status.Capacity.Storage().String(),
		}
	}

	ingress, err := firstIngress(ctx, h.clientset, kls.Namespace, listOptions)
	if err != nil {
		result.Ingress = &models.ExplainIngress{
			Error: errors.Wrap(err, "failed to get ingress").Error(),
		}
	} else {
		hosts := make([]string, 0)
		for _, rule := range ingress.Spec.Rules {
			hosts = append(hosts, rule.Host)
		}

		result.Ingress = &models.ExplainIngress{
			Hosts:        hosts,
			IngressClass: *ingress.Spec.IngressClassName,
		}
	}

	return apiService.NewServiceExplainOK().WithPayload(result)
}

func firstPod(ctx context.Context, clientset kubernetes.Interface, ns string, listOptions metav1.ListOptions) (*v1.Pod, error) {
	objects, err := clientset.CoreV1().Pods(ns).List(ctx, listOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error listing pod objects")
	}

	if len(objects.Items) < 1 {
		return nil, errors.New("no pod is available")
	}
	obj := objects.Items[0]
	return &obj, nil
}

func firstPVC(ctx context.Context, clientset kubernetes.Interface, ns string, listOptions metav1.ListOptions) (*v1.PersistentVolumeClaim, error) {
	objects, err := clientset.CoreV1().PersistentVolumeClaims(ns).List(ctx, listOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error listing pvc objects")
	}

	if len(objects.Items) < 1 {
		return nil, errors.New("no pvc is available")
	}
	obj := objects.Items[0]
	return &obj, nil
}

func firstIngress(ctx context.Context, clientset kubernetes.Interface, ns string, listOptions metav1.ListOptions) (*v12.Ingress, error) {
	objects, err := clientset.NetworkingV1().Ingresses(ns).List(ctx, listOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error listing ingresses objects")
	}

	if len(objects.Items) < 1 {
		return nil, errors.New("no ingress is available")
	}
	obj := objects.Items[0]
	return &obj, nil
}

func containerStatus(state v1.ContainerState) string {
	switch {
	case state.Waiting != nil:
		return "waiting"
	case state.Running != nil:
		return "running"
	case state.Terminated != nil:
		return "terminated"
	default:
		return "unknown"
	}
}
