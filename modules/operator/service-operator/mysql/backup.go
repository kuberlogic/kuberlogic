package mysql

import (
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	"github.com/kuberlogic/kuberlogic/modules/operator/service-operator/base"
	"github.com/kuberlogic/kuberlogic/modules/operator/util"
	v1 "k8s.io/api/core/v1"
)

const (
	backupImage = "backup-mysql"
	backupTag   = "latest"
)

type Backup struct {
	base.BaseBackup
	Cluster *Mysql
}

func (p *Backup) SetBackupImage() {
	p.Image = util.GetKuberlogicImage(backupImage, backupTag)
}

func (p *Backup) SetBackupEnv(cm *kuberlogicv1.KuberLogicBackupSchedule) {
	env := []v1.EnvVar{
		{
			Name:  "SCOPE",
			Value: cm.Name,
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
		// mysql env vars
		{
			Name:      "MYSQL_PASSWORD",
			ValueFrom: util.FromSecret(p.Cluster.Operator.Spec.SecretName, passwordField),
		},
		{
			Name:  "DATABASE",
			Value: cm.Spec.Database,
		},
	}
	env = append(env, util.BucketVariables(cm.Spec.SecretName)...)
	env = append(env, util.S3Credentials(cm.Spec.SecretName)...)
	env = append(env, util.SentryEnv()...)
	p.EnvVar = env
}
