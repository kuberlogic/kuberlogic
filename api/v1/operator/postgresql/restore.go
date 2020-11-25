package postgresql

import (
	"fmt"
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	"gitlab.com/cloudmanaged/operator/api/v1/operator/backup"
	"gitlab.com/cloudmanaged/operator/api/v1/operator/util"
	v12 "k8s.io/api/core/v1"
)

const (
	restoreImage = "cloudmanaged-restore-postgresql"
	restoreTag   = "latest"
)

type Restore struct {
	backup.BaseRestore
	Cluster Postgres
}

func (p *Restore) SetRestoreImage() {
	p.Image = util.GetImage(restoreImage, restoreTag)
}

func (p *Restore) SetRestoreEnv(cm *cloudlinuxv1.CloudManagedRestore) {
	pgDataSecret := fmt.Sprintf("%s.%s.credentials", postgreSuperUser,
		p.Cluster.Operator.Name)

	env := []v12.EnvVar{
		{
			Name:  "SCOPE",
			Value: p.Cluster.Operator.Name,
		},
		{
			Name:      "CLUSTER_NAME_LABEL",
			ValueFrom: util.FromConfigMap(operatorConfigMap, "cluster_name_label"),
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &v12.EnvVarSource{
				FieldRef: &v12.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
		// Postgres env vars
		{
			Name:  "PG_VERSION",
			Value: p.Cluster.Operator.Spec.PostgresqlParam.PgVersion,
		},
		{
			Name:  "PGPORT",
			Value: "5432",
		},
		{
			Name:      "PGUSER",
			ValueFrom: util.FromSecret(pgDataSecret, "username"),
		},
		{
			Name:  "PGDATABASE",
			Value: postgreSuperUser,
		},
		{
			Name:  "PGSSLMODE",
			Value: "require",
		},
		{
			Name:      "PGPASSWORD",
			ValueFrom: util.FromSecret(pgDataSecret, "password"),
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
