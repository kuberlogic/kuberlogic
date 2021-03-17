package base

import (
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/util"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type BaseRestore struct {
	Job batchv1.Job

	// set in the specific operator
	Image  string
	EnvVar []corev1.EnvVar
}

func (r *BaseRestore) Init(crb *kuberlogicv1.KuberLogicBackupRestore) {
	r.Job = r.New(crb)
}

func (r *BaseRestore) InitFrom(job *batchv1.Job) {
	r.Job = *job
}

func (r *BaseRestore) CurrentStatus() *kuberlogicv1.KuberLogicBackupRestoreStatus {
	s := new(kuberlogicv1.KuberLogicBackupRestoreStatus)

	if len(r.Job.Status.Conditions) > 0 {
		lastCondition := r.Job.Status.Conditions[len(r.Job.Status.Conditions)-1]
		switch lastCondition.Type {
		case batchv1.JobComplete:
			s.Status = kuberlogicv1.BackupSuccessStatus
		case batchv1.JobFailed:
			s.Status = kuberlogicv1.BackupFailedStatus
		}
		s.CompletionTime = lastCondition.LastTransitionTime.Format(time.RFC3339)
	} else {
		if r.Job.Status.Active > 0 {
			s.Status = kuberlogicv1.BackupRunningStatus
		}
	}

	return s
}

func (r *BaseRestore) GetJob() *batchv1.Job {
	return &r.Job
}

func (r *BaseRestore) New(crb *kuberlogicv1.KuberLogicBackupRestore) batchv1.Job {
	return r.NewJob(crb.Name, crb.Namespace)
}

func (r *BaseRestore) NewJob(name, ns string) batchv1.Job {
	var backOffLimit int32 = 2

	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backOffLimit,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            name,
							Image:           r.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env:             r.EnvVar,
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: util.GetImagePullSecret()},
					},
					RestartPolicy: corev1.RestartPolicyOnFailure,
				},
			},
		},
	}
}
