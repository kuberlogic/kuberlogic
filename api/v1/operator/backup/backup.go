package backup

import (
	kuberlogicv1 "gitlab.com/cloudmanaged/operator/api/v1"
	"gitlab.com/cloudmanaged/operator/api/v1/operator/util"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sort"
)

type BaseBackup struct {
	CronJob v1beta1.CronJob

	Image  string
	EnvVar []corev1.EnvVar
}

func (p *BaseBackup) CurrentStatus(ev v1.JobList) string {
	// sort in reverse order
	sort.SliceStable(ev.Items, func(i, j int) bool {
		first, second := ev.Items[i], ev.Items[j]
		return second.Status.StartTime.Before(first.Status.StartTime)
	})

	if len(ev.Items) > 0 {
		lastJob := ev.Items[0]
		if len(lastJob.Status.Conditions) > 0 {
			lastCondition := lastJob.Status.Conditions[len(lastJob.Status.Conditions)-1]
			switch lastCondition.Type {
			case v1.JobComplete:
				return kuberlogicv1.BackupSuccessStatus
			case v1.JobFailed:
				return kuberlogicv1.BackupFailedStatus
			}
		} else {
			if lastJob.Status.Active > 0 {
				return kuberlogicv1.BackupRunningStatus
			}
		}
	}
	return kuberlogicv1.BackupUnknownStatus
}

func (p *BaseBackup) NewCronJob(name, ns, schedule string) v1beta1.CronJob {
	labels := map[string]string{
		"backup-name": name,
	}
	var backOffLimit int32 = 2

	return v1beta1.CronJob{
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
							Containers: []corev1.Container{
								{
									Name:            name,
									Image:           p.Image,
									ImagePullPolicy: corev1.PullIfNotPresent,
									Env:             p.EnvVar,
								},
							},
							ImagePullSecrets: []corev1.LocalObjectReference{
								{Name: util.GetImagePullSecret()},
							},
							RestartPolicy: corev1.RestartPolicyOnFailure,
						},
					},
				},
			},
		},
	}
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
