/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package base

import (
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	"github.com/kuberlogic/kuberlogic/modules/operator/util"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

type BaseBackup struct {
	CronJob batchv1beta1.CronJob

	Image          string
	ServiceAccount string
	EnvVar         []corev1.EnvVar
}

func (p *BaseBackup) IsFailed(j *batchv1.Job) bool {
	for _, c := range j.Status.Conditions {
		if c.Type == batchv1.JobFailed {
			return true
		}
	}
	return false
}

func (p BaseBackup) IsFinished(j *batchv1.Job) bool {
	for _, c := range j.Status.Conditions {
		if c.Type == batchv1.JobFailed || c.Type == batchv1.JobComplete {
			return true
		}
	}
	return false
}

func (p *BaseBackup) IsRunning(j *batchv1.Job) bool {
	return j.Status.Active > 0
}

func (p *BaseBackup) NewCronJob(name, ns, schedule string) batchv1beta1.CronJob {
	labels := map[string]string{
		"backup-name": name,
	}
	var backOffLimit int32 = 2

	c := batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: batchv1beta1.CronJobSpec{
			Schedule:          schedule,
			ConcurrencyPolicy: batchv1beta1.ForbidConcurrent,
			JobTemplate: batchv1beta1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: batchv1.JobSpec{
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

func (p *BaseBackup) GetCronJob() *batchv1beta1.CronJob {
	return &p.CronJob
}

func (p *BaseBackup) New(backup *kuberlogicv1.KuberLogicBackupSchedule) batchv1beta1.CronJob {
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

func (p *BaseBackup) InitFrom(job *batchv1beta1.CronJob) {
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

func (p *BaseBackup) SetImage(repo, image, version string) {
	p.Image = repo + "/" + image + ":" + version
}
