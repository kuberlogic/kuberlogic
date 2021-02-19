package mysql

import (
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	"gitlab.com/cloudmanaged/operator/api/v1/operator/backup"
	"gitlab.com/cloudmanaged/operator/api/v1/operator/util"
	v1 "k8s.io/api/core/v1"
)

const (
	restoreImage = "cloudmanaged-restore-mysql"
	restoreTag   = "latest"
)

type Restore struct {
	backup.BaseRestore
	Cluster Mysql
}

func (p *Restore) SetRestoreImage() {
	p.Image = util.GetImage(restoreImage, restoreTag)
}

func (p *Restore) SetRestoreEnv(cm *cloudlinuxv1.KuberLogicBackupRestore) {
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
			Name:  "PATH_TO_BACKUP",
			Value: cm.Spec.Backup,
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
