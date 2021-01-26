package postgresql

import (
	"fmt"
	postgresv1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	"gitlab.com/cloudmanaged/operator/api/v1/operator/util"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	image   = "postgresql"
	version = "12.1.5"

	teamId = "cloudmanaged"

	postgresRoleKey     = "spilo-role"
	postgresRoleReplica = "replica"
	postgresRoleMaster  = "master"

	postgresPodLabelKey = "cluster-name"

	postgresPodDefaultKey = "application"
	postgresPodDefaultVal = "spilo"

	postgresMainContainer = "postgres"

	postgresPort = 5432
)

type Postgres struct {
	Operator postgresv1.Postgresql
}

func (p *Postgres) AsRuntimeObject() runtime.Object {
	return &p.Operator
}

func (p *Postgres) AsMetaObject() metav1.Object {
	return &p.Operator
}

func (p *Postgres) InitFrom(o runtime.Object) {
	p.Operator = *o.(*postgresv1.Postgresql)
}

func (p *Postgres) Init(cm *cloudlinuxv1.CloudManaged) {
	loadBalancersEnabled := true

	p.Operator = postgresv1.Postgresql{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", teamId, cm.Name),
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
										Name: fmt.Sprintf("%s.%s.credentials", "cloudmanaged", cm.Name),
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
										Name: fmt.Sprintf("%s.%s.credentials", "cloudmanaged", cm.Name),
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

func (p *Postgres) GetDefaults() cloudlinuxv1.Defaults {
	return cloudlinuxv1.Defaults{
		VolumeSize: cloudlinuxv1.DefaultVolumeSize,
		Resources:  cloudlinuxv1.DefaultResources,
		Version:    version,
		User:       cloudlinuxv1.DefaultUser,
	}
}

func (p *Postgres) Update(cm *cloudlinuxv1.CloudManaged) {
	p.setReplica(cm)
	p.setResources(cm)
	p.setVolumeSize(cm)
	p.setImage(cm)
	p.setAdvancedConf(cm)
}

func (p *Postgres) setReplica(cm *cloudlinuxv1.CloudManaged) {
	p.Operator.Spec.NumberOfInstances = cm.Spec.Replicas
}

func (p *Postgres) setResources(cm *cloudlinuxv1.CloudManaged) {
	op := &p.Operator.Spec.Resources
	cmr := &cm.Spec.Resources

	op.ResourceLimits.CPU, op.ResourceLimits.Memory = cmr.Limits.Cpu().String(), cmr.Limits.Memory().String()
	op.ResourceRequests.CPU, op.ResourceRequests.Memory = cmr.Requests.Cpu().String(), cmr.Requests.Memory().String()
}

func (p *Postgres) setVolumeSize(cm *cloudlinuxv1.CloudManaged) {
	p.Operator.Spec.Volume.Size = cm.Spec.VolumeSize
}

func (p *Postgres) setImage(cm *cloudlinuxv1.CloudManaged) {
	p.Operator.Spec.DockerImage = util.GetImage(image, cm.Spec.Version)
}

func (p *Postgres) setAdvancedConf(cm *cloudlinuxv1.CloudManaged) {
	if p.Operator.Spec.PostgresqlParam.Parameters == nil {
		p.Operator.Spec.PostgresqlParam.Parameters = make(map[string]string)
	}

	for k, v := range cm.Spec.AdvancedConf {
		p.Operator.Spec.PostgresqlParam.Parameters[k] = v
	}
}

func (p *Postgres) IsEqual(cm *cloudlinuxv1.CloudManaged) bool {
	return p.isEqualReplica(cm) &&
		p.isEqualResources(cm) &&
		p.isEqualVolumeSize(cm) &&
		p.isEqualImage(cm) &&
		p.isEqualAdvancedConf(cm)
}

func (p *Postgres) isEqualReplica(cm *cloudlinuxv1.CloudManaged) bool {
	return p.Operator.Spec.NumberOfInstances == cm.Spec.Replicas
}

func (p *Postgres) isEqualResources(cm *cloudlinuxv1.CloudManaged) bool {
	op := p.Operator.Spec.Resources
	cmr := cm.Spec.Resources
	return op.ResourceLimits.CPU == cmr.Limits.Cpu().String() &&
		op.ResourceLimits.Memory == cmr.Limits.Memory().String() &&
		op.ResourceRequests.CPU == cmr.Requests.Cpu().String() &&
		op.ResourceRequests.Memory == cmr.Requests.Memory().String()
}

func (p *Postgres) isEqualVolumeSize(cm *cloudlinuxv1.CloudManaged) bool {
	return p.Operator.Spec.Volume.Size == cm.Spec.VolumeSize
}

func (p *Postgres) isEqualImage(cm *cloudlinuxv1.CloudManaged) bool {
	return p.Operator.Spec.DockerImage == util.GetImage(image, cm.Spec.Version)
}

func (p *Postgres) isEqualAdvancedConf(cm *cloudlinuxv1.CloudManaged) bool {
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
		return cloudlinuxv1.ClusterNotReadyStatus
	case postgresv1.ClusterStatusAddFailed, postgresv1.ClusterStatusUpdateFailed, postgresv1.ClusterStatusSyncFailed, postgresv1.ClusterStatusInvalid:
		return cloudlinuxv1.ClusterFailedStatus
	case postgresv1.ClusterStatusRunning:
		return cloudlinuxv1.ClusterOkStatus
	default:
		return cloudlinuxv1.ClusterUnknownStatus
	}
}

func (p *Postgres) GetPodReplicaSelector() map[string]string {
	return map[string]string{postgresRoleKey: postgresRoleReplica,
		postgresPodLabelKey:   p.Operator.ObjectMeta.Name,
		postgresPodDefaultKey: postgresPodDefaultVal,
	}
}

func (p *Postgres) GetPodMasterSelector() map[string]string {
	return map[string]string{postgresRoleKey: postgresRoleMaster,
		postgresPodLabelKey:   p.Operator.ObjectMeta.Name,
		postgresPodDefaultKey: postgresPodDefaultVal,
	}
}

func (p *Postgres) GetMasterService() string {
	return fmt.Sprintf("%s", p.Operator.ObjectMeta.Name)
}

func (p *Postgres) GetReplicaService() string {
	return fmt.Sprintf("%s-repl", p.Operator.ObjectMeta.Name)
}

func (p *Postgres) GetAccessPort() int {
	return postgresPort
}

func (p *Postgres) GetMainPodContainer() string {
	return postgresMainContainer
}

func (p *Postgres) GetDefaultConnectionPassword() (secret, passwordField string) {
	return fmt.Sprintf("%s.%s.credentials", cloudlinuxv1.DefaultUser, p.Operator.ObjectMeta.Name), "password"
}

func (p *Postgres) GetCredentialsSecret() (*apiv1.Secret, error) {
	return nil, nil
}
