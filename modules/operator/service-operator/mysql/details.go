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

package mysql

import (
	"fmt"
	util2 "github.com/bitpoke/mysql-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	mysqlRoleKey     = "role"
	mysqlRoleReplica = "replica"
	mysqlRoleMaster  = "master"

	mysqlPodLabelKey   = "mysql.presslabs.org/cluster"
	mysqlMainContainer = "mysql"

	mysqlPort = 3306

	passwordField = "ROOT_PASSWORD"
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

func (d *InternalDetails) GetDefaultConnectionPassword() (string, string) {
	return d.Cluster.Operator.Spec.SecretName, passwordField
}

func (d *InternalDetails) GetDefaultConnectionUser() string {
	return masterUser
}

func (d *InternalDetails) GetCredentialsSecret() (*corev1.Secret, error) {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      genCredentialsSecretName(d.Cluster.Operator.ObjectMeta.Name),
			Namespace: d.Cluster.Operator.ObjectMeta.Namespace,
		},
		StringData: map[string]string{
			passwordField: util2.RandomString(15),
		},
	}, nil
}
