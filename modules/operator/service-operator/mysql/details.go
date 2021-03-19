package mysql

import (
	"fmt"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	util2 "github.com/presslabs/mysql-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type InternalDetails struct {
	Cluster *Mysql
}

func (d *InternalDetails) GetPodReplicaSelector() map[string]string {
	return map[string]string{
		mysqlRoleKey:     mysqlRoleReplica,
		mysqlPodLabelKey: d.Cluster.Operator.ObjectMeta.Name,
	}
}

func (d *InternalDetails) GetPodMasterSelector() map[string]string {
	return map[string]string{
		mysqlRoleKey:     mysqlRoleMaster,
		mysqlPodLabelKey: d.Cluster.Operator.ObjectMeta.Name,
	}
}

func (d *InternalDetails) GetMasterService() string {
	return fmt.Sprintf("%s-mysql-master", d.Cluster.Operator.ObjectMeta.Name)
}

func (d *InternalDetails) GetReplicaService() string {
	return fmt.Sprintf("%s-mysql-replicas", d.Cluster.Operator.ObjectMeta.Name)
}

func (d *InternalDetails) GetAccessPort() int {
	return mysqlPort
}

func (d *InternalDetails) GetMainPodContainer() string {
	return mysqlMainContainer
}

func (d *InternalDetails) GetDefaultConnectionPassword() (secret, passwordField string) {
	return d.Cluster.Operator.Spec.SecretName, "PASSWORD"
}

func (d *InternalDetails) GetCredentialsSecret() (*corev1.Secret, error) {
	rootPassword := util2.RandomString(15)
	userName := kuberlogicv1.DefaultUser
	userPassword := util2.RandomString(15)

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      genCredentialsSecretName(d.Cluster.Operator.ObjectMeta.Name),
			Namespace: d.Cluster.Operator.ObjectMeta.Namespace,
		},
		StringData: map[string]string{
			"ROOT_PASSWORD": rootPassword,
			"USER":          userName,
			"PASSWORD":      userPassword,
		},
	}, nil
}
