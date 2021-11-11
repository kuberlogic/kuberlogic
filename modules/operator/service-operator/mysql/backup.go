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
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	"github.com/kuberlogic/kuberlogic/modules/operator/service-operator/base"
	"github.com/kuberlogic/kuberlogic/modules/operator/util"
	v1 "k8s.io/api/core/v1"
)

const (
	backupImage = "backup-mysql"
)

type Backup struct {
	base.BaseBackup
	Cluster *Mysql
}

func (p *Backup) SetBackupImage(repo, version string) {
	p.SetImage(repo, backupImage, version)
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
