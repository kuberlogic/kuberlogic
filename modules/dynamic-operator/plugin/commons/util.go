/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package commons

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func ResponseFromObject(object client.Object, gvk schema.GroupVersionKind) *PluginResponse {
	o, err := ToUnstructured(object, gvk)
	if err != nil {
		return &PluginResponse{
			Err: err.Error(),
		}
	}
	return &PluginResponse{
		Object: o,
	}

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
