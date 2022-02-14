/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugin

import (
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

func (p *PostgresqlService) SetLogger(logger hclog.Logger) {
	p.logger = logger
}

func (p *PostgresqlService) Default() *commons.PluginResponseDefault {
	p.logger.Debug("call Default")

	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(
		&postgresv1.Resources{
			ResourceRequests: postgresv1.ResourceDescription{
				CPU:    "100m",
				Memory: "128Mi",
			},
			ResourceLimits: postgresv1.ResourceDescription{
				CPU:    "250m",
				Memory: "256Mi",
			},
		})
	if err != nil {
		p.logger.Error("cannot convert object", "err", err)
		return &commons.PluginResponseDefault{
			Error: err.Error(),
		}
	}

	p.logger.Debug("content", "c", content)
	return &commons.PluginResponseDefault{
		Replicas:   1,
		VolumeSize: "1Gi",
		Version:    "13",
		Parameters: map[string]interface{}{
			"resources": content,
		},
	}
}

func (p *PostgresqlService) ValidateCreate(req commons.PluginRequest) *commons.PluginResponse {
	p.logger.Debug("call ValidateCreate")
	return &commons.PluginResponse{
		Error: "",
	}
}

func (p *PostgresqlService) ValidateUpdate(req commons.PluginRequest) *commons.PluginResponse {
	p.logger.Debug("call ValidateUpdate")
	return &commons.PluginResponse{
		Error: "",
	}
}

func (p *PostgresqlService) ValidateDelete(req commons.PluginRequest) *commons.PluginResponse {
	p.logger.Debug("call ValidateDelete")
	return &commons.PluginResponse{
		Error: "",
	}
}

func (p *PostgresqlService) Type() *commons.PluginResponse {
	return commons.ResponseFromObject(&postgresv1.Postgresql{}, gvk())
}

func (p *PostgresqlService) Empty(req commons.PluginRequest) *commons.PluginResponse {
	p.logger.Debug("call Empty")

	return commons.ResponseFromObject(
		&postgresv1.Postgresql{
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name,
				Namespace: req.Namespace,
			},
		}, gvk())
}

func (p *PostgresqlService) merge(object *postgresv1.Postgresql, req commons.PluginRequest) error {
	object.Spec.NumberOfInstances = req.Replicas
	object.Spec.Volume.Size = req.VolumeSize
	object.Spec.PgVersion = req.Version

	for k, v := range req.Parameters {
		switch k {
		case "resources":
			result := &postgresv1.Resources{}
			err := commons.FromUnstructured(v.(map[string]interface{}), result)
			if err != nil {
				return errors.New(fmt.Sprintf("cannot convert %v to postgresv1.Resources", v))
			}
			object.Spec.Resources = *result
		default:
			// unknown parameter
			p.logger.Error("unknown parameter", "key", k, "value", v)
		}
	}
	return nil
}

func (p *PostgresqlService) IsReady() bool {
	switch p.service.Status.PostgresClusterStatus {
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

func (p *PostgresqlService) Status(req commons.PluginRequest) *commons.PluginResponse {
	p.logger.Debug("call Status")

	object := &postgresv1.Postgresql{}
	err := commons.FromUnstructured(req.Object.UnstructuredContent(), object)
	if err != nil {
		p.logger.Error(err.Error())
		return &commons.PluginResponse{
			Error: err.Error(),
		}
	}

	isReady := p.IsReady()
	p.logger.Info("isReady", "result", isReady)
	return &commons.PluginResponse{
		IsReady: isReady,
	}
}

func (p *PostgresqlService) ForUpdate(req commons.PluginRequest) *commons.PluginResponse {
	p.logger.Debug("call ForUpdate")
	object := &postgresv1.Postgresql{}
	err := commons.FromUnstructured(req.Object.UnstructuredContent(), object)
	if err != nil {
		p.logger.Error(err.Error())
		return &commons.PluginResponse{
			Error: err.Error(),
		}
	}

	err = p.merge(object, req)
	if err != nil {
		p.logger.Error(err.Error())
		return &commons.PluginResponse{
			Error: err.Error(),
		}

	}
	p.logger.Info("ForUpdate", "object", object)
	return commons.ResponseFromObject(object, gvk())
}

func (p *PostgresqlService) ForCreate(req commons.PluginRequest) *commons.PluginResponse {
	p.logger.Debug("call ForCreate")
	object := p.Init(req)
	err := p.merge(object, req)
	if err != nil {
		p.logger.Error(err.Error())
		return &commons.PluginResponse{
			Error: err.Error(),
		}

	}

	return commons.ResponseFromObject(object, gvk())
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
