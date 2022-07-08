package kuberlogicservice_env

import (
	"context"
	"encoding/json"
	"fmt"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/controllers/backuprestore"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-service-pod,mutating=true,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod.kuberlogic.com,admissionReviewVersions=v1,sideEffects=NoneOnDryRun,failurePolicy=ignore

type ServicePodWebhook struct {
	Client  client.Client
	decoder *admission.Decoder
}

const (
	ignorePauseLabel = "kuberlogic.com/ignore-pause"
)

var (
	pausedAffinityTerm = corev1.PodAffinityTerm{
		LabelSelector: &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      ignorePauseLabel,
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"true"},
				},
			},
		},
		TopologyKey: "kubernetes.io/hostname",
	}
)

func (m *ServicePodWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}
	err := m.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// check if kls is paused
	kls := &kuberlogiccomv1alpha1.KuberLogicService{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Namespace,
		},
	}

	if err := m.Client.Get(ctx, client.ObjectKeyFromObject(kls), kls); err != nil {
		return admission.Allowed(fmt.Sprintf("failed to find or get corresponding kls: %s", err))
	}

	// whitelist backup/restore pod
	if pod.Name != backuprestore.ResticBackupPodName {
		restoreRunning, _ := kls.RestoreRunning()
		backupRunning, _ := kls.BackupRunning()
		if restoreRunning || backupRunning || kls.Paused() {
			pod.Spec.Affinity = &corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						pausedAffinityTerm,
					},
				},
			}
			marshaledPod, err := json.Marshal(pod)
			if err != nil {
				return admission.Errored(http.StatusInternalServerError, err)
			}
			return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
		}
	}
	return admission.Allowed("pod is allowed")
}

func (m *ServicePodWebhook) InjectDecoder(d *admission.Decoder) error {
	m.decoder = d
	return nil
}
