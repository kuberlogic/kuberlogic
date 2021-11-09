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

package postgresql

import (
	"fmt"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	"github.com/kuberlogic/kuberlogic/modules/operator/service-operator/interfaces"
	platformOp "github.com/kuberlogic/kuberlogic/modules/operator/service-operator/postgresql/platform"
	"github.com/kuberlogic/kuberlogic/modules/operator/util"
	"github.com/pkg/errors"
	postgresv1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	image  = "postgresql"
	tag    = "spilo-13-2.0-p6"
	teamId = "kuberlogic"
)

type Postgres struct {
	Operator         postgresv1.Postgresql
	platformOperator interfaces.PlatformOperator
}

func (p *Postgres) GetBackupSchedule() interfaces.BackupSchedule {
	return &Backup{
		Cluster: p,
	}
}

func (p *Postgres) GetBackupRestore() interfaces.BackupRestore {
	return &Restore{
		Cluster: p,
	}
}

func (p *Postgres) GetInternalDetails() interfaces.InternalDetails {
	return &InternalDetails{
		Cluster: p,
	}
}

func (p *Postgres) GetSession(kls *kuberlogicv1.KuberLogicService, client kubernetes.Interface, db string) (interfaces.Session, error) {
	return NewSession(p, kls, client, db)
}

func (p *Postgres) AsRuntimeObject() runtime.Object {
	return &p.Operator
}

func (p *Postgres) AsMetaObject() metav1.Object {
	return &p.Operator
}

func (p *Postgres) AsClientObject() client.Object {
	return &p.Operator
}

func (p *Postgres) Name(kls *kuberlogicv1.KuberLogicService) string {
	return fmt.Sprintf("%s-%s", teamId, kls.Name)
}

func (p *Postgres) InitFrom(o runtime.Object) {
	p.Operator = *o.(*postgresv1.Postgresql)
}

func (p *Postgres) Init(kls *kuberlogicv1.KuberLogicService, platform string) {
	loadBalancersEnabled := true

	name := p.Name(kls)
	defaultUserCredentialsSecret := genUserCredentialsSecretName(teamId, name)

	p.Operator = postgresv1.Postgresql{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: kls.Namespace,
		},
		Spec: postgresv1.PostgresSpec{
			TeamID:                    teamId,
			EnableMasterLoadBalancer:  &loadBalancersEnabled,
			EnableReplicaLoadBalancer: &loadBalancersEnabled,
			Users: map[string]postgresv1.UserFlags{
				// required user like teamId name with necessary credentials
				teamId: {"superuser", "createdb"},
			},
			DockerImage: util.GetKuberlogicImage(image, tag),
			PostgresqlParam: postgresv1.PostgresqlParam{
				PgVersion: kls.Spec.Version,
				Parameters: map[string]string{
					"shared_buffers":  "32MB",
					"max_connections": "10",
					"log_statement":   "all",
				},
			},
			Patroni: postgresv1.Patroni{
				InitDB: map[string]string{
					"encoding":       "UTF8",
					"locale":         "en_US.UTF-8",
					"data-checksums": "true",
				},
				//PgHba:                []string{"hostssl all all 0.0.0.0/0 md5", "host    all all 0.0.0.0/0 md5"},
				TTL:                  30,
				LoopWait:             10,
				RetryTimeout:         10,
				MaximumLagOnFailover: 33554432,
				Slots:                map[string]map[string]string{},
			},
			PodAnnotations: map[string]string{
				"monitoring.cloudlinux.com/scrape": "true",
				"monitoring.cloudlinux.com/port":   "9187",
			},
			Sidecars: []postgresv1.Sidecar{
				{
					Name:        "postgres-exporter",
					DockerImage: "quay.io/kuberlogic/bitnami-postgres-exporter:0.8.0",
					Ports: []apiv1.ContainerPort{
						{
							Name:          "metrics",
							ContainerPort: 9187,
							Protocol:      apiv1.ProtocolTCP,
						},
					},
					Env: []apiv1.EnvVar{
						{
							Name:  "DATA_SOURCE_URI",
							Value: "127.0.0.1/postgres?sslmode=disable",
						},
						{
							Name: "DATA_SOURCE_USER",
							ValueFrom: &apiv1.EnvVarSource{
								SecretKeyRef: &apiv1.SecretKeySelector{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: defaultUserCredentialsSecret,
									},
									Key: "username",
								},
							},
						},
						{
							Name: "DATA_SOURCE_PASS",
							ValueFrom: &apiv1.EnvVarSource{
								SecretKeyRef: &apiv1.SecretKeySelector{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: defaultUserCredentialsSecret,
									},
									Key: passwordField,
								},
							},
						},
					},
				},
			},
		},
	}
	p.platformOperator = platformOp.NewPlatformOperator(&p.Operator, platform)
}

func (p *Postgres) Update(kls *kuberlogicv1.KuberLogicService) error {
	p.setReplica(kls)
	p.setResources(kls)
	p.setVolumeSize(kls)
	p.setVersion(kls)
	p.setAdvancedConf(kls)

	if err := p.platformOperator.SetAllowedIPs(kuberlogicv1.DefaultAllowedIPs); err != nil {
		return errors.Wrap(err, "error applying platforms changes")
	}
	return nil
}

func (p *Postgres) setReplica(kls *kuberlogicv1.KuberLogicService) {
	p.Operator.Spec.NumberOfInstances = kls.Spec.Replicas
}

func (p *Postgres) setResources(kls *kuberlogicv1.KuberLogicService) {
	op := &p.Operator.Spec.Resources
	klsr := &kls.Spec.Resources

	op.ResourceLimits.CPU, op.ResourceLimits.Memory = klsr.Limits.Cpu().String(), klsr.Limits.Memory().String()
	op.ResourceRequests.CPU, op.ResourceRequests.Memory = klsr.Requests.Cpu().String(), klsr.Requests.Memory().String()
}

func (p *Postgres) setVolumeSize(kls *kuberlogicv1.KuberLogicService) {
	p.Operator.Spec.Volume.Size = kls.Spec.VolumeSize
}

func (p *Postgres) setVersion(kls *kuberlogicv1.KuberLogicService) {
	p.Operator.Spec.PostgresqlParam.PgVersion = kls.Spec.Version
}

func (p *Postgres) setAdvancedConf(kls *kuberlogicv1.KuberLogicService) {
	if p.Operator.Spec.PostgresqlParam.Parameters == nil {
		p.Operator.Spec.PostgresqlParam.Parameters = make(map[string]string)
	}

	for k, v := range kls.Spec.AdvancedConf {
		p.Operator.Spec.PostgresqlParam.Parameters[k] = v
	}
}

func (p *Postgres) IsReady() (bool, string) {
	switch p.Operator.Status.PostgresClusterStatus {
	case postgresv1.ClusterStatusCreating, postgresv1.ClusterStatusUpdating, postgresv1.ClusterStatusUnknown:
		return false, kuberlogicv1.ClusterNotReadyStatus
	case postgresv1.ClusterStatusAddFailed, postgresv1.ClusterStatusUpdateFailed, postgresv1.ClusterStatusSyncFailed, postgresv1.ClusterStatusInvalid:
		return false, kuberlogicv1.ClusterFailedStatus
	case postgresv1.ClusterStatusRunning:
		return true, kuberlogicv1.ClusterOkStatus
	default:
		return false, kuberlogicv1.ClusterUnknownStatus
	}
}

func genUserCredentialsSecretName(user, cluster string) string {
	return fmt.Sprintf("%s.%s.credentials", user, cluster)
}
