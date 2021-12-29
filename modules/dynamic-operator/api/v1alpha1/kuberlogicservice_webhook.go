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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var kuberlogicservicelog = logf.Log.WithName("kuberlogicservice-resource")

func (r *KuberLogicService) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-kuberlogic-com-kuberlogic-com-v1alpha1-kuberlogicservice,mutating=true,failurePolicy=fail,sideEffects=None,groups=kuberlogic.com.kuberlogic.com,resources=kuberlogicservices,verbs=create;update,versions=v1alpha1,name=mkuberlogicservice.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &KuberLogicService{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *KuberLogicService) Default() {
	kuberlogicservicelog.Info("default", "name", r.Name)

}

//+kubebuilder:webhook:path=/validate-kuberlogic-com-kuberlogic-com-v1alpha1-kuberlogicservice,mutating=false,failurePolicy=fail,sideEffects=None,groups=kuberlogic.com.kuberlogic.com,resources=kuberlogicservices,verbs=create;update;delete,versions=v1alpha1,name=vkuberlogicservice.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &KuberLogicService{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *KuberLogicService) ValidateCreate() error {
	kuberlogicservicelog.Info("validate create", "name", r.Name)

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *KuberLogicService) ValidateUpdate(old runtime.Object) error {
	kuberlogicservicelog.Info("validate update", "name", r.Name)

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *KuberLogicService) ValidateDelete() error {
	kuberlogicservicelog.Info("validate delete", "name", r.Name)

	return nil
}
