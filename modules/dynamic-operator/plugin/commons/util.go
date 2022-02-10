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
	"github.com/hashicorp/go-plugin"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
)

func FromUnstructured(u map[string]interface{}, obj interface{}) error {
	return runtime.DefaultUnstructuredConverter.FromUnstructured(u, obj)
}

func ToUnstructured(obj interface{}, gvk schema.GroupVersionKind) (*unstructured.Unstructured, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	u := &unstructured.Unstructured{Object: content}
	u.SetGroupVersionKind(gvk)
	return u, nil
}

func ServePlugin(name string, pl PluginService) {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	pl.SetLogger(logger)
	var pluginMap = map[string]plugin.Plugin{
		name: &Plugin{Impl: pl},
	}

	logger.Debug("starting the plugin", "type", name)
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins:         pluginMap,
	})
}
