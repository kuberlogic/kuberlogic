package base

import (
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	"github.com/kuberlogic/kuberlogic/modules/operator/util"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

type BaseBackup struct {
	CronJob v1beta1.CronJob

	Image          string
	ServiceAccount string
	EnvVar         []corev1.EnvVar
}

func (p *BaseBackup) IsSuccessful(j *v1.Job) bool {
	for _, c := range j.Status.Conditions {
		if c.Type == v1.JobComplete {
			return true
		}
	}
	return false
}

func (p *BaseBackup) IsRunning(j *v1.Job) bool {
	return j.Status.Active > 0
}

func (p *BaseBackup) NewCronJob(name, ns, schedule string) v1beta1.CronJob {
	labels := map[string]string{
		"backup-name": name,
	}
	var backOffLimit int32 = 2

	c := v1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: v1beta1.CronJobSpec{
			Schedule:          schedule,
			ConcurrencyPolicy: v1beta1.ForbidConcurrent,
			JobTemplate: v1beta1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.JobSpec{
					BackoffLimit: &backOffLimit,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							ServiceAccountName: p.ServiceAccount,
							Containers: []corev1.Container{
								{
									Name:            name,
									Image:           p.Image,
									ImagePullPolicy: corev1.PullIfNotPresent,
									Env:             p.EnvVar,
								},
							},
							RestartPolicy: corev1.RestartPolicyOnFailure,
						},
					},
				},
			},
		},
	}

	if util.GetKuberlogicRepoPullSecret() != "" {
		c.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{Name: util.GetKuberlogicRepoPullSecret()},
		}
	}
	return c
}

func (p *BaseBackup) GetCronJob() *v1beta1.CronJob {
	return &p.CronJob
}

func (p *BaseBackup) New(backup *kuberlogicv1.KuberLogicBackupSchedule) v1beta1.CronJob {
	return p.NewCronJob(
		backup.Name,
		backup.Namespace,
		backup.Spec.Schedule,
	)
}

func (p *BaseBackup) GetBackupImage() string {
	panic("not implemented error")
}

func (p *BaseBackup) GetBackupEnv(secret string) []corev1.EnvVar {
	panic("not implemented error")
}

func (p *BaseBackup) Init(cm *kuberlogicv1.KuberLogicBackupSchedule) {
	p.CronJob = p.New(cm)
}

func (p *BaseBackup) InitFrom(job *v1beta1.CronJob) {
	p.CronJob = *job
}

func (p *BaseBackup) IsEqual(cm *kuberlogicv1.KuberLogicBackupSchedule) bool {
	return p.IsEqualSchedule(cm) &&
		p.IsEqualTemplate(cm)
}

func (p *BaseBackup) IsEqualSchedule(cm *kuberlogicv1.KuberLogicBackupSchedule) bool {
	return reflect.DeepEqual(cm.Spec.Schedule, p.CronJob.Spec.Schedule)
}

func (p *BaseBackup) IsEqualTemplate(cm *kuberlogicv1.KuberLogicBackupSchedule) bool {
	return reflect.DeepEqual(
		p.New(cm).Spec.JobTemplate,
		p.CronJob.Spec.JobTemplate,
	)
}

func (p *BaseBackup) Update(cm *kuberlogicv1.KuberLogicBackupSchedule) {
	p.UpdateSchedule(cm)
	p.UpdateTemplate(cm)
}

func (p *BaseBackup) UpdateSchedule(cm *kuberlogicv1.KuberLogicBackupSchedule) {
	p.CronJob.Spec.Schedule = cm.Spec.Schedule
}

func (p *BaseBackup) UpdateTemplate(cm *kuberlogicv1.KuberLogicBackupSchedule) {
	p.CronJob.Spec.JobTemplate = p.New(cm).Spec.JobTemplate
}

func (p *BaseBackup) SetServiceAccount(name string) {
	p.ServiceAccount = name
}
