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
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type kuberLogicServiceTypeApiRef struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

type KuberlogicServiceTypeParam struct {
	Path         string  `json:"path"`
	DefaultValue v1.JSON `json:"defaultValue,omitempty"`
	Type         string  `json:"type,omitempty"`
	Conversion   string  `json:"conversion,omitempty"`
}

type KuberLogicServiceTypeConditions struct {
	Path           string `json:"path"`
	ReadyCondition string `json:"readyCondition"`
	ReadyValue     string `json:"readyValue"`
}

type KuberlogicServiceTypeStatusRef struct {
	Conditions *KuberLogicServiceTypeConditions `json:"conditions,omitempty"`
}

type KuberLogicServiceTypeSpec struct {
	Type string `json:"type"`

	Api         kuberLogicServiceTypeApiRef           `json:"api"`
	SpecRef     map[string]KuberlogicServiceTypeParam `json:"specRef"`
	StatusRef   KuberlogicServiceTypeStatusRef        `json:"statusRef"`
	DefaultSpec v1.JSON                               `json:"defaultSpec"`
}

// KuberLogicServiceTypeStatus defines the observed state of KuberLogicServiceType
type KuberLogicServiceTypeStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=klst,categories=kuberlogic,scope=Cluster
// +kubebuilder:subresource:status

type KuberLogicServiceType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KuberLogicServiceTypeSpec   `json:"spec,omitempty"`
	Status KuberLogicServiceTypeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KuberLogicServiceTypeList contains a list of KuberLogicServiceType
type KuberLogicServiceTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KuberLogicServiceType `json:"items"`
}

func (k KuberLogicServiceType) ServiceGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   k.Spec.Api.Group,
		Version: k.Spec.Api.Version,
		Kind:    k.Spec.Api.Kind,
	}
}

func init() {
	SchemeBuilder.Register(&KuberLogicServiceType{}, &KuberLogicServiceTypeList{})
}
