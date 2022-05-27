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

package plugin

import (
	compose "github.com/compose-spec/compose-go/types"
	"github.com/hashicorp/go-hclog"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	pluginCompose "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugins/docker-compose/plugin/compose/compose"
	"strings"
)

type DockerComposeService struct {
	logger hclog.Logger
	spec   *compose.Project
}

func (d *DockerComposeService) SetLogger(logger hclog.Logger) {
	d.logger = logger
}

func (d *DockerComposeService) Convert(req commons.PluginRequest) *commons.PluginResponse {
	res := &commons.PluginResponse{}

	composeObjects := pluginCompose.NewComposeModel(d.spec, d.logger)
	objects, err := composeObjects.Reconcile(&req)
	if err != nil {
		d.logger.Error(err.Error(), "error reconciling cluster objects")
		res.Err = err.Error()
		return res
	}

	for _, item := range objects {
		for gvk, object := range item {
			_ = res.AddObject(object, gvk)
		}
	}
	return res
}

func (d *DockerComposeService) Status(req commons.PluginRequest) *commons.PluginResponseStatus {
	status := &commons.PluginResponseStatus{
		IsReady: false,
	}

	composeObjects := pluginCompose.NewComposeModel(d.spec, d.logger)
	ready, err := composeObjects.Ready(&req)
	if err != nil {
		d.logger.Error(err.Error(), "error checking for readiness")
		status.Err = err.Error()
	}
	status.IsReady = ready

	return status
}

func (d *DockerComposeService) Types() *commons.PluginResponse {
	res := &commons.PluginResponse{}

	composeObjects := pluginCompose.NewComposeModel(d.spec, d.logger)
	types := composeObjects.Types()

	// we need to filter duplicates first
	for _, item := range types {
		for gvk, object := range item {
			_ = res.AddObject(object, gvk)
		}
	}

	return res
}

func (d *DockerComposeService) Default() *commons.PluginResponseDefault {
	return &commons.PluginResponseDefault{
		Replicas:   1,
		VolumeSize: "1Gi",
	}
}

func (d *DockerComposeService) ValidateCreate(req commons.PluginRequest) *commons.PluginResponseValidation {
	var validateErrors []string
	if req.Replicas != 1 {
		validateErrors = append(validateErrors, "only 1 replica can be set")
	}
	return &commons.PluginResponseValidation{
		Err: strings.Join(validateErrors, ","),
	}
}

func (d *DockerComposeService) ValidateUpdate(req commons.PluginRequest) *commons.PluginResponseValidation {
	var validateErrors []string
	if req.Replicas != 1 {
		validateErrors = append(validateErrors, "only 1 replica can be set")
	}
	return &commons.PluginResponseValidation{
		Err: strings.Join(validateErrors, ","),
	}
}

func (d *DockerComposeService) ValidateDelete(req commons.PluginRequest) *commons.PluginResponseValidation {
	return &commons.PluginResponseValidation{}
}

func NewDockerComposeServicePlugin(composeProject *compose.Project) *DockerComposeService {
	return &DockerComposeService{
		spec: composeProject,
	}
}
