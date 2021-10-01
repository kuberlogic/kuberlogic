package base

import (
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/util"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BaseRestore struct {
	Job batchv1.Job

	// set in the specific operator
	Image          string
	EnvVar         []corev1.EnvVar
	ServiceAccount string
}

func (r *BaseRestore) Init(crb *kuberlogicv1.KuberLogicBackupRestore) {
	r.Job = r.New(crb)
}

func (r *BaseRestore) InitFrom(job *batchv1.Job) {
	r.Job = *job
}

func (r *BaseRestore) IsSuccessful() bool {
	for _, c := range r.Job.Status.Conditions {
		if c.Type == batchv1.JobComplete {
			return true
		}
	}
	return false
}

func (r *BaseRestore) IsFailed() bool {
	for _, c := range r.Job.Status.Conditions {
		if c.Type == batchv1.JobFailed {
			return true
		}
	}
	return false
}

func (r *BaseRestore) IsRunning() bool {
	return r.Job.Status.Active > 0
}

func (r *BaseRestore) GetJob() *batchv1.Job {
	return &r.Job
}

func (r *BaseRestore) SetServiceAccount(name string) {
	r.ServiceAccount = name
}

func (r *BaseRestore) New(crb *kuberlogicv1.KuberLogicBackupRestore) batchv1.Job {
	return r.NewJob(crb.Name, crb.Namespace)
}

func (r *BaseRestore) NewJob(name, ns string) batchv1.Job {
	var backOffLimit int32 = 2

	j := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backOffLimit,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: r.ServiceAccount,
					Containers: []corev1.Container{
						{
							Name:            name,
							Image:           r.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env:             r.EnvVar,
						},
					},
					RestartPolicy: corev1.RestartPolicyOnFailure,
				},
			},
		},
	}

	if util.GetKuberlogicRepoPullSecret() != "" {
		j.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{Name: util.GetKuberlogicRepoPullSecret()},
		}
	}
	return j
}
