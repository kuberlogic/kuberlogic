package mysql

import (
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	"gitlab.com/cloudmanaged/operator/api/v1/operator/backup"
	"gitlab.com/cloudmanaged/operator/api/v1/operator/util"
	v1 "k8s.io/api/core/v1"
)

const (
	backupImage = "cloudmanaged-backup-mysql"
	backupTag   = "latest"
)

type Backup struct {
	backup.BaseBackup
	Cluster Mysql
}

func (p *Backup) SetBackupImage() {
	p.Image = util.GetImage(backupImage, backupTag)
}

func (p *Backup) SetBackupEnv(cm *cloudlinuxv1.CloudManagedBackup) {
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
	}
	env = append(env, util.BucketVariables(cm.Spec.SecretName)...)
	env = append(env, util.S3Credentials(cm.Spec.SecretName)...)
	p.EnvVar = env
}
