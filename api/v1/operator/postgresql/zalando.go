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
			Name:      cm.Name,
			Namespace: cm.Namespace,
		},
		Spec: postgresv1.PostgresSpec{
			TeamID:                    "cloudmanaged",
			EnableMasterLoadBalancer:  &loadBalancersEnabled,
			EnableReplicaLoadBalancer: &loadBalancersEnabled,
			Users: map[string]postgresv1.UserFlags{
				// required user like teamId name with necessary credentials
				"cloudmanaged": {"superuser", "createdb"},
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
	}
}

func (p *Postgres) Update(cm *cloudlinuxv1.CloudManaged) {
	p.setReplica(cm)
	p.setResources(cm)
	p.setVolumeSize(cm)
	p.setImage(cm)
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

func (p *Postgres) IsEqual(cm *cloudlinuxv1.CloudManaged) bool {
	return p.isEqualReplica(cm) &&
		p.isEqualResources(cm) &&
		p.isEqualVolumeSize(cm) &&
		p.isEqualImage(cm)
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
