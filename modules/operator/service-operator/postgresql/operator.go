package postgresql

import (
	"fmt"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	service_operator "github.com/kuberlogic/operator/modules/operator/service-operator"
	"github.com/kuberlogic/operator/modules/operator/service-operator/base"
	"github.com/kuberlogic/operator/modules/operator/util"
	postgresv1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	image   = "postgresql"
	version = "12.1.5"
	teamId  = "kuberlogic"
)

type Postgres struct {
	base.BaseOperator
	Operator *postgresv1.Postgresql
}

func (p *Postgres) GetBackupSchedule() service_operator.BackupSchedule {
	return &Backup{
		Cluster: p,
	}
}

func (p *Postgres) GetBackupRestore() service_operator.BackupRestore {
	return &Restore{
		Cluster: p,
	}
}

func (p *Postgres) GetInternalDetails() service_operator.InternalDetails {
	return &InternalDetails{
		Cluster: p,
	}
}

func (p *Postgres) Name(cm *kuberlogicv1.KuberLogicService) string {
	return fmt.Sprintf("%s-%s", teamId, cm.Name)
}

func (p *Postgres) InitFrom(o runtime.Object) {
	p.Operator = o.(*postgresv1.Postgresql)
}

func (p *Postgres) Init(cm *kuberlogicv1.KuberLogicService) {
	loadBalancersEnabled := true

	name := p.Name(cm)
	defaultUserCredentialsSecret := genUserCredentialsSecretName(teamId, name)

	p.Operator = &postgresv1.Postgresql{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cm.Namespace,
		},
		Spec: postgresv1.PostgresSpec{
			TeamID:                    teamId,
			EnableMasterLoadBalancer:  &loadBalancersEnabled,
			EnableReplicaLoadBalancer: &loadBalancersEnabled,
			Users: map[string]postgresv1.UserFlags{
				// required user like teamId name with necessary credentials
				teamId: {"superuser", "createdb"},
			},
			PostgresqlParam: postgresv1.PostgresqlParam{
				PgVersion: "12",
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
				PgHba:                []string{"hostssl all all 0.0.0.0/0 md5", "host    all all 0.0.0.0/0 md5"},
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
					DockerImage: "bitnami/postgres-exporter:0.8.0",
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
									Key: "password",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (p *Postgres) GetDefaults() kuberlogicv1.Defaults {
	return kuberlogicv1.Defaults{
		VolumeSize: kuberlogicv1.DefaultVolumeSize,
		Resources:  kuberlogicv1.DefaultResources,
		Version:    version,
	}
}

func (p *Postgres) Update(cm *kuberlogicv1.KuberLogicService) {
	p.setReplica(cm)
	p.setResources(cm)
	p.setVolumeSize(cm)
	p.setImage(cm)
	p.setAdvancedConf(cm)
}

func (p *Postgres) setReplica(cm *kuberlogicv1.KuberLogicService) {
	p.Operator.Spec.NumberOfInstances = cm.Spec.Replicas
}

func (p *Postgres) setResources(cm *kuberlogicv1.KuberLogicService) {
	op := &p.Operator.Spec.Resources
	cmr := &cm.Spec.Resources

	op.ResourceLimits.CPU, op.ResourceLimits.Memory = cmr.Limits.Cpu().String(), cmr.Limits.Memory().String()
	op.ResourceRequests.CPU, op.ResourceRequests.Memory = cmr.Requests.Cpu().String(), cmr.Requests.Memory().String()
}

func (p *Postgres) setVolumeSize(cm *kuberlogicv1.KuberLogicService) {
	p.Operator.Spec.Volume.Size = cm.Spec.VolumeSize
}

func (p *Postgres) setImage(cm *kuberlogicv1.KuberLogicService) {
	p.Operator.Spec.DockerImage = util.GetImage(image, cm.Spec.Version)
}

func (p *Postgres) setAdvancedConf(cm *kuberlogicv1.KuberLogicService) {
	if p.Operator.Spec.PostgresqlParam.Parameters == nil {
		p.Operator.Spec.PostgresqlParam.Parameters = make(map[string]string)
	}

	for k, v := range cm.Spec.AdvancedConf {
		p.Operator.Spec.PostgresqlParam.Parameters[k] = v
	}
}

func (p *Postgres) IsEqual(cm *kuberlogicv1.KuberLogicService) bool {
	return p.isEqualReplica(cm) &&
		p.isEqualResources(cm) &&
		p.isEqualVolumeSize(cm) &&
		p.isEqualImage(cm) &&
		p.isEqualAdvancedConf(cm)
}

func (p *Postgres) isEqualReplica(cm *kuberlogicv1.KuberLogicService) bool {
	return p.Operator.Spec.NumberOfInstances == cm.Spec.Replicas
}

func (p *Postgres) isEqualResources(cm *kuberlogicv1.KuberLogicService) bool {
	op := p.Operator.Spec.Resources
	cmr := cm.Spec.Resources
	return op.ResourceLimits.CPU == cmr.Limits.Cpu().String() &&
		op.ResourceLimits.Memory == cmr.Limits.Memory().String() &&
		op.ResourceRequests.CPU == cmr.Requests.Cpu().String() &&
		op.ResourceRequests.Memory == cmr.Requests.Memory().String()
}

func (p *Postgres) isEqualVolumeSize(cm *kuberlogicv1.KuberLogicService) bool {
	return p.Operator.Spec.Volume.Size == cm.Spec.VolumeSize
}

func (p *Postgres) isEqualImage(cm *kuberlogicv1.KuberLogicService) bool {
	return p.Operator.Spec.DockerImage == util.GetImage(image, cm.Spec.Version)
}

func (p *Postgres) isEqualAdvancedConf(cm *kuberlogicv1.KuberLogicService) bool {
	for k, v := range cm.Spec.AdvancedConf {
		if val, ok := p.Operator.Spec.PostgresqlParam.Parameters[k]; !ok {
			return false
		} else if val != v {
			return false
		}
	}
	return true
}

func (p *Postgres) CurrentStatus() string {
	switch p.Operator.Status.PostgresClusterStatus {
	case postgresv1.ClusterStatusCreating, postgresv1.ClusterStatusUpdating, postgresv1.ClusterStatusUnknown:
		return kuberlogicv1.ClusterNotReadyStatus
	case postgresv1.ClusterStatusAddFailed, postgresv1.ClusterStatusUpdateFailed, postgresv1.ClusterStatusSyncFailed, postgresv1.ClusterStatusInvalid:
		return kuberlogicv1.ClusterFailedStatus
	case postgresv1.ClusterStatusRunning:
		return kuberlogicv1.ClusterOkStatus
	default:
		return kuberlogicv1.ClusterUnknownStatus
	}
}

func genUserCredentialsSecretName(user, cluster string) string {
	return fmt.Sprintf("%s.%s.credentials", user, cluster)
}
