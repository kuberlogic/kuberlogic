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

package v1alpha1

import (
	"encoding/json"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var log = ctrl.Log.WithName("kuberlogicservice-webhook")

var pluginInstances map[string]commons.PluginService

func (r *KuberLogicService) SetupWebhookWithManager(mgr ctrl.Manager, plugins map[string]commons.PluginService) error {
	pluginInstances = plugins
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-kuberlogic-com-v1alpha1-kuberlogicservice,mutating=true,failurePolicy=fail,sideEffects=None,groups=kuberlogic.com,resources=kuberlogicservices,verbs=create;update,versions=v1alpha1,name=mkuberlogicservice.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &KuberLogicService{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *KuberLogicService) Default() {
	log.Info("default", "name", r.Name)

	plugin, ok := pluginInstances[r.Spec.Type]
	if !ok {
		log.Info("Plugin is not loaded", "type", r.Spec.Type)
		return
	}

	resp := plugin.Default()
	if resp.Error != nil {
		log.Error(resp.Error, "error rpc call 'Default'")
		return
	}
	if r.Spec.Replicas == 0 {
		r.Spec.Replicas = resp.Replicas
	}
	if r.Spec.VolumeSize == "" {
		r.Spec.VolumeSize = resp.VolumeSize
	}
	if r.Spec.Version == "" {
		r.Spec.Version = resp.Version
	}

	spec := make(map[string]interface{}, 0)
	if len(r.Spec.Advanced.Raw) > 0 {
		if err := json.Unmarshal(r.Spec.Advanced.Raw, &spec); err != nil {
			log.Error(err, "error unmarshalling spec")
			return
		}
	}

	found := false
	for k, defaultValue := range resp.Parameters {
		_, exists := spec[k]
		if !exists {
			spec[k] = defaultValue
			found = true
		}
	}
	if found {
		bytes, err := json.Marshal(spec)
		if err != nil {
			log.Error(err, "error marshalling spec")
			return
		}
		r.Spec.Advanced.Raw = bytes
	}
}

//+kubebuilder:webhook:path=/validate-kuberlogic-com-v1alpha1-kuberlogicservice,mutating=false,failurePolicy=fail,sideEffects=None,groups=kuberlogic.com,resources=kuberlogicservices,verbs=create;update;delete,versions=v1alpha1,name=vkuberlogicservice.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &KuberLogicService{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *KuberLogicService) ValidateCreate() error {
	log.Info("validate create", "name", r.Name)

	plugin, ok := pluginInstances[r.Spec.Type]
	if !ok {
		err := errors.New("Plugin is not loaded")
		log.Info(err.Error(), "type", r.Spec.Type)
		return err
	}

	req, err := makeRequest(r)
	if err != nil {
		return err
	}
	return plugin.ValidateCreate(*req).Error
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *KuberLogicService) ValidateUpdate(old runtime.Object) error {
	log.Info("validate update", "name", r.Name)

	plugin, ok := pluginInstances[r.Spec.Type]
	if !ok {
		err := errors.New("Plugin is not loaded")
		log.Info(err.Error(), "type", r.Spec.Type)
		return err
	}

	req, err := makeRequest(r)
	if err != nil {
		return err
	}
	return plugin.ValidateUpdate(*req).Error
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *KuberLogicService) ValidateDelete() error {
	log.Info("validate delete", "name", r.Name)

	plugin, ok := pluginInstances[r.Spec.Type]
	if !ok {
		err := errors.New("Plugin is not loaded")
		log.Info(err.Error(), "type", r.Spec.Type)
		return err
	}

	req, err := makeRequest(r)
	if err != nil {
		return err
	}
	return plugin.ValidateDelete(*req).Error
}

func makeRequest(kls *KuberLogicService) (*commons.PluginRequest, error) {

	spec := make(map[string]interface{}, 0)
	if len(kls.Spec.Advanced.Raw) > 0 {
		if err := json.Unmarshal(kls.Spec.Advanced.Raw, &spec); err != nil {
			log.Error(err, "error unmarshalling spec")
			return nil, err
		}
	}
	return &commons.PluginRequest{
		Name:       kls.Name,
		Namespace:  kls.Namespace,
		Replicas:   kls.Spec.Replicas,
		VolumeSize: kls.Spec.VolumeSize,
		Version:    kls.Spec.Version,
		Parameters: spec,
	}, nil
}
