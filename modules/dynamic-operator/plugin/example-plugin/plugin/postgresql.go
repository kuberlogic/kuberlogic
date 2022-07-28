/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package plugin

import (
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"

	"github.com/hashicorp/go-hclog"
	postgresv1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
)

const (
	image  = "postgresql"
	tag    = "spilo-13-2.0-p6"
	teamId = "kuberlogic"

	passwordField = "password"
	imageRepo     = "quay.io/kuberlogic"
)

var _ commons.PluginService = &PostgresqlService{}

type PostgresqlService struct {
	logger  hclog.Logger
	service postgresv1.Postgresql
}

func NewPostgresqlService(logger hclog.Logger) *PostgresqlService {
	return &PostgresqlService{
		logger: logger,
	}
}

func (p *PostgresqlService) Default() *commons.PluginResponseDefault {
	p.logger.Debug("call Default")

	v := &commons.PluginResponseDefault{
		Replicas:   1,
		VolumeSize: "1Gi",
		Version:    "13",
	}
	_ = v.SetLimits(&apiv1.ResourceList{
		"cpu":    resource.MustParse("250m"),
		"memory": resource.MustParse("256Mi"),
	})

	p.logger.Debug("limits", "res", v.Limits)
	return v
}

func (p *PostgresqlService) ValidateCreate(req commons.PluginRequest) *commons.PluginResponseValidation {
	p.logger.Debug("call ValidateCreate", "ns", req.Namespace, "name", req.Name)
	return &commons.PluginResponseValidation{}
}

func (p *PostgresqlService) ValidateUpdate(req commons.PluginRequest) *commons.PluginResponseValidation {
	p.logger.Debug("call ValidateUpdate", "ns", req.Namespace, "name", req.Name)
	return &commons.PluginResponseValidation{}
}

func (p *PostgresqlService) ValidateDelete(req commons.PluginRequest) *commons.PluginResponseValidation {
	p.logger.Debug("call ValidateDelete", "ns", req.Namespace, "name", req.Name)
	return &commons.PluginResponseValidation{}
}

func (p *PostgresqlService) Types() *commons.PluginResponse {
	p.logger.Debug("call Type")
	return commons.ResponseFromObject(&postgresv1.Postgresql{}, gvk(), "", commons.TCPproto)
}

func (p *PostgresqlService) merge(object *postgresv1.Postgresql, req commons.PluginRequest) error {
	object.Spec.NumberOfInstances = req.Replicas
	object.Spec.Volume.Size = req.VolumeSize
	object.Spec.PgVersion = req.Version

	from, err := req.GetLimits()
	if err != nil {
		return err
	}
	to := &object.Spec.Resources
	to.ResourceLimits.CPU, to.ResourceLimits.Memory = from.Cpu().String(), from.Memory().String()
	to.ResourceRequests.CPU, to.ResourceRequests.Memory = "100m", "128Mi" // default values

	for k, v := range req.Parameters {
		switch k {
		//case "resources":
		//	result := &postgresv1.Resources{}
		//	err := commons.FromUnstructured(v.(map[string]interface{}), result)
		//	if err != nil {
		//		return errors.New(fmt.Sprintf("cannot convert %v to postgresv1.Resources", v))
		//	}
		//	object.Spec.Resources = *result
		default:
			// unknown parameter
			p.logger.Error("unknown parameter", "key", k, "value", v)
		}
	}
	return nil
}

func (p *PostgresqlService) IsReady(service *postgresv1.Postgresql) bool {
	switch service.Status.PostgresClusterStatus {
	case postgresv1.ClusterStatusCreating, postgresv1.ClusterStatusUpdating, postgresv1.ClusterStatusUnknown:
		return false
	case postgresv1.ClusterStatusAddFailed, postgresv1.ClusterStatusUpdateFailed, postgresv1.ClusterStatusSyncFailed, postgresv1.ClusterStatusInvalid:
		return false
	case postgresv1.ClusterStatusRunning:
		return true
	default:
		return false
	}
}

func (p *PostgresqlService) Status(req commons.PluginRequest) *commons.PluginResponseStatus {
	p.logger.Debug("call Status", "ns", req.Namespace, "name", req.Name)

	existingObjs := req.GetObjects()
	object := &postgresv1.Postgresql{}
	err := commons.FromUnstructured(existingObjs[0].UnstructuredContent(), object)
	if err != nil {
		p.logger.Error(err.Error())
		return &commons.PluginResponseStatus{
			Err: err.Error(),
		}
	}

	isReady := p.IsReady(object)
	p.logger.Info("isReady", "result", isReady)
	return &commons.PluginResponseStatus{
		IsReady: isReady,
	}
}

func (p *PostgresqlService) Convert(req commons.PluginRequest) *commons.PluginResponse {
	p.logger.Debug("call Convert", "ns", req.Namespace, "name", req.Name)

	var object *postgresv1.Postgresql
	existingObjs := req.GetObjects()
	if existingObjs != nil && existingObjs[0] != nil {
		// using existing object
		object = &postgresv1.Postgresql{}
		err := commons.FromUnstructured(existingObjs[0].UnstructuredContent(), object)
		if err != nil {
			p.logger.Error(err.Error())
			return &commons.PluginResponse{
				Err: err.Error(),
			}
		}
	} else {
		// creating a new one
		object = p.Init(req)
	}

	err := p.merge(object, req)
	if err != nil {
		p.logger.Error(err.Error())
		return &commons.PluginResponse{
			Err: err.Error(),
		}

	}
	p.logger.Info("Convert", "object", object)

	return commons.ResponseFromObject(object, gvk(), p.service.GetName(), commons.TCPproto)
}

func (p *PostgresqlService) Init(req commons.PluginRequest) *postgresv1.Postgresql {
	loadBalancersEnabled := true
	defaultUserCredentialsSecret := genUserCredentialsSecretName(teamId, req.Name)

	return &postgresv1.Postgresql{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Spec: postgresv1.PostgresSpec{
			TeamID:                    teamId,
			EnableMasterLoadBalancer:  &loadBalancersEnabled,
			EnableReplicaLoadBalancer: &loadBalancersEnabled,
			Users: map[string]postgresv1.UserFlags{
				// required user like teamId name with necessary credentials
				teamId: {"superuser", "createdb"},
			},
			DockerImage: getImage(image, tag),
			PostgresqlParam: postgresv1.PostgresqlParam{
				PgVersion: req.Version,
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
				"monitoring.kuberlogic.com/scrape": "true",
				"monitoring.kuberlogic.com/port":   "9187",
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
}

func getImage(base, v string) string {
	return fmt.Sprintf("%s/%s:%s", strings.TrimSuffix(imageRepo, "/"), base, v)
}

func genUserCredentialsSecretName(user, cluster string) string {
	return fmt.Sprintf("%s.%s.credentials", user, cluster)
}

func gvk() schema.GroupVersionKind {
	return postgresv1.SchemeGroupVersion.WithKind(postgresv1.PostgresCRDResourceKind)
}
