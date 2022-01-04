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
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	errorsApi "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var log = ctrl.Log.WithName("kuberlogicservice-webhook")
var cl client.Client

func (r *KuberLogicService) SetupWebhookWithManager(mgr ctrl.Manager) error {
	cl = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-kuberlogic-com-v1alpha1-kuberlogicservice,mutating=true,failurePolicy=fail,sideEffects=None,groups=kuberlogic.com,resources=kuberlogicservices,verbs=create;update,versions=v1alpha1,name=mkuberlogicservice.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &KuberLogicService{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *KuberLogicService) Default() {
	log.Info("default", "name", r.Name)

	klst, spec, err := r.fetchKuberLogicServiceType()
	if err != nil {
		log.Error(err, "cannot fetch KuberLogicServiceType resource")
	}

	setDefaults(spec, klst)
	bytes, err := json.Marshal(spec)
	if err != nil {
		log.Error(err, "cannot marshal spec")
	}
	r.Spec.Raw = bytes
}

//+kubebuilder:webhook:path=/validate-kuberlogic-com-v1alpha1-kuberlogicservice,mutating=false,failurePolicy=fail,sideEffects=None,groups=kuberlogic.com,resources=kuberlogicservices,verbs=create;update;delete,versions=v1alpha1,name=vkuberlogicservice.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &KuberLogicService{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *KuberLogicService) ValidateCreate() error {
	log.Info("validate create", "name", r.Name)
	return r.ValidateByType()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *KuberLogicService) ValidateUpdate(old runtime.Object) error {
	log.Info("validate update", "name", r.Name)
	return r.ValidateByType()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *KuberLogicService) ValidateDelete() error {
	log.Info("validate delete", "name", r.Name)

	return nil
}

func (r *KuberLogicService) fetchKuberLogicServiceType() (
	*KuberLogicServiceType,
	*map[string]interface{},
	error,
) {
	ctx := context.TODO()
	spec := make(map[string]interface{}, 0)
	if err := json.Unmarshal(r.Spec.Raw, &spec); err != nil {
		log.Error(err, "cannot unmarshal spec")
		return nil, nil, err
	}

	klst := &KuberLogicServiceType{}
	if err := cl.Get(ctx, types.NamespacedName{
		Name:      r.GetServiceType(spec),
		Namespace: r.Namespace,
	}, klst); err != nil {
		log.Error(err, "cannot fetch linked KuberLogicServiceType resource")
		return nil, nil, err
	}
	return klst, &spec, nil
}

func (r *KuberLogicService) ValidateByType() error {
	var allErrs field.ErrorList
	log.Info("validate by type", "name", r.Name)

	klst, spec, err := r.fetchKuberLogicServiceType()
	if err != nil {
		log.Error(err, "cannot fetch KuberLogicServiceType resource")
		return err
	}

	for k, _ := range *spec {
		if k == "type" { // skip specific "type" parameter
			continue
		}
		_, ok := klst.Spec.SpecRef[k]
		if !ok {
			err := errors.New("key is not found in type spec")
			log.Error(err, "validating error")
			allErrs = append(allErrs, field.Invalid(
				field.NewPath("spec").Child("SpecRef").Child(k),
				nil,
				err.Error(),
			))
		}
	}

	if len(allErrs) > 0 {
		return errorsApi.NewInvalid(
			r.GroupVersionKind().GroupKind(),
			r.Name,
			allErrs)
	}
	return nil
}

func setDefaults(
	spec *map[string]interface{},
	klst *KuberLogicServiceType,
) {
	for k, typeValue := range klst.Spec.SpecRef {
		_, ok := (*spec)[k]
		if !ok {
			log.Info("key is not found in type spec, using default",
				"key", k, "value", typeValue.DefaultValue)

			var v interface{}
			if err := json.Unmarshal(typeValue.DefaultValue.Raw, &v); err != nil {
				log.Error(err, "error unmarshaling default value")
			}
			(*spec)[k] = v
		}
	}
}
