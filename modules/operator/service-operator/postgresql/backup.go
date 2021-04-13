package postgresql

import (
	"fmt"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/service-operator/base"
	"github.com/kuberlogic/operator/modules/operator/util"
	v1 "k8s.io/api/core/v1"
)

const (
	backupImage = "backup-postgresql"
	backupTag   = "latest"

	operatorConfigMap = "cm-postgres-operator"
	postgreSuperUser  = "postgres" // TODO: Could be grabbed from config map ^
)

type Backup struct {
	base.BaseBackup
	Cluster *Postgres
}

func (p *Backup) SetBackupImage() {
	p.Image = util.GetImage(backupImage, backupTag)
}

func (p *Backup) SetBackupEnv(cm *kuberlogicv1.KuberLogicBackupSchedule) {
	pgDataSecret := fmt.Sprintf("%s.%s.credentials", postgreSuperUser,
		p.Cluster.Operator.Name)

	env := []v1.EnvVar{
		{
			Name:  "SCOPE",
			Value: cm.Name,
		},
		{
			Name:      "CLUSTER_NAME_LABEL",
			ValueFrom: util.FromConfigMap(operatorConfigMap, "cluster_name_label"),
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
			Name:  "DATABASE",
			Value: cm.Spec.Database,
		},
	}
	env = append(env, util.BucketVariables(cm.Spec.SecretName)...)
	env = append(env, util.S3Credentials(cm.Spec.SecretName)...)
	p.EnvVar = env
}
