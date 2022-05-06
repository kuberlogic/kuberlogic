/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
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
	if resp.Error() != nil {
		log.Error(resp.Error(), "error rpc call 'Default'")
		return
	}
	if r.Spec.VolumeSize == "" {
		r.Spec.VolumeSize = resp.VolumeSize
	}
	if r.Spec.Version == "" {
		r.Spec.Version = resp.Version
	}

	if r.Spec.Limits == nil {
		r.Spec.Limits = *resp.GetLimits()
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
	if err = plugin.ValidateCreate(*req).Error(); err != nil {
		return err
	}
	return nil
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

	if err = plugin.ValidateUpdate(*req).Error(); err != nil {
		return err
	}
	return nil
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

	if err := plugin.ValidateDelete(*req).Error(); err != nil {
		return err
	}
	return nil
}

func makeRequest(kls *KuberLogicService) (*commons.PluginRequest, error) {
	spec := make(map[string]interface{}, 0)
	if len(kls.Spec.Advanced.Raw) > 0 {
		if err := json.Unmarshal(kls.Spec.Advanced.Raw, &spec); err != nil {
			log.Error(err, "error unmarshalling spec")
			return nil, err
		}
	}
	req := &commons.PluginRequest{
		Name:       kls.Name,
		Namespace:  kls.Namespace,
		Replicas:   kls.Spec.Replicas,
		VolumeSize: kls.Spec.VolumeSize,
		Version:    kls.Spec.Version,
		Parameters: spec,
	}
	err := req.SetLimits(&kls.Spec.Limits)
	if err != nil {
		return nil, err
	}

	return req, nil
}