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

package commons

import (
	"github.com/hashicorp/go-hclog"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// PluginService is the interface that we're exposing as a plugin.
type PluginService interface {
	SetLogger(logger hclog.Logger)
	Empty(req PluginRequest) *PluginResponse
	ForCreate(req PluginRequest) *PluginResponse
	ForUpdate(req PluginRequest) *PluginResponse
	Status(req PluginRequest) *PluginResponse
	Type() *PluginResponse

	Default() *PluginResponseDefault
	ValidateCreate(req PluginRequest) *PluginResponse
	ValidateUpdate(req PluginRequest) *PluginResponse
	ValidateDelete(req PluginRequest) *PluginResponse
}

type PluginRequestEmpty struct{}
type PluginRequest struct {
	Name      string
	Namespace string

	Replicas   int32
	VolumeSize string
	Version    string

	Parameters map[string]interface{}
	Object     *unstructured.Unstructured
}

type PluginResponse struct {
	Object  *unstructured.Unstructured
	IsReady bool
	Error   string
}

type PluginResponseDefault struct {
	Replicas   int32
	VolumeSize string
	Version    string
	Parameters map[string]interface{}

	Error string
}
