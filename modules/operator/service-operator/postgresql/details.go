package postgresql

import (
	"fmt"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	apiv1 "k8s.io/api/core/v1"
)

const (
	postgresRoleKey       = "spilo-role"
	postgresRoleReplica   = "replica"
	postgresRoleMaster    = "master"
	postgresPodLabelKey   = "cluster-name"
	postgresPodDefaultKey = "application"
	postgresPodDefaultVal = "spilo"
	postgresMainContainer = "postgres"
	postgresPort          = 5432
)

type InternalDetails struct {
	Cluster *Postgres
}

func (d *InternalDetails) GetPodReplicaSelector() map[string]string {
	return map[string]string{postgresRoleKey: postgresRoleReplica,
		postgresPodLabelKey:   d.Cluster.Operator.ObjectMeta.Name,
		postgresPodDefaultKey: postgresPodDefaultVal,
	}
}

func (d *InternalDetails) GetPodMasterSelector() map[string]string {
	return map[string]string{postgresRoleKey: postgresRoleMaster,
		postgresPodLabelKey:   d.Cluster.Operator.ObjectMeta.Name,
		postgresPodDefaultKey: postgresPodDefaultVal,
	}
}

func (d *InternalDetails) GetMasterService() string {
	return fmt.Sprintf("%s", d.Cluster.Operator.ObjectMeta.Name)
}

func (d *InternalDetails) GetReplicaService() string {
	return fmt.Sprintf("%s-repl", d.Cluster.Operator.ObjectMeta.Name)
}

func (d *InternalDetails) GetAccessPort() int {
	return postgresPort
}

func (d *InternalDetails) GetMainPodContainer() string {
	return postgresMainContainer
}

func (d *InternalDetails) GetDefaultConnectionPassword() (secret, passwordField string) {
	return genUserCredentialsSecretName(kuberlogicv1.DefaultUser, d.Cluster.Operator.ObjectMeta.Name), "password"
}

func (d *InternalDetails) GetCredentialsSecret() (*apiv1.Secret, error) {
	return nil, nil
}
