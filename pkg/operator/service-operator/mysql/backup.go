package mysql

import (
	kuberlogicv1 "github.com/kuberlogic/operator/pkg/operator/api/v1"
	"github.com/kuberlogic/operator/pkg/operator/service-operator/base"
	"github.com/kuberlogic/operator/pkg/operator/util"
	v1 "k8s.io/api/core/v1"
)

const (
	backupImage = "backup-mysql"
	backupTag   = "latest"
)

type Backup struct {
	base.BaseBackup
	Cluster Mysql
}

func (p *Backup) SetBackupImage() {
	p.Image = util.GetImage(backupImage, backupTag)
}

func (p *Backup) SetBackupEnv(cm *kuberlogicv1.KuberLogicBackupSchedule) {
	env := []v1.EnvVar{
		{
			Name:  "SCOPE",
			Value: p.Cluster.Operator.Name,
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
			ValueFrom: util.FromSecret(p.Cluster.Operator.Spec.SecretName, "ROOT_PASSWORD"),
		},
		{
			Name:  "DATABASE",
			Value: cm.Spec.Database,
		},
	}
	env = append(env, util.BucketVariables(cm.Spec.SecretName)...)
	env = append(env, util.S3Credentials(cm.Spec.SecretName)...)
	p.EnvVar = env
}
