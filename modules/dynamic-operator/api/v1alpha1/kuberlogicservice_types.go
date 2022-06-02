/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	v11 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	readyCondType        = "Ready"
	clusterUnknownStatus = "Unknown"
)

// KuberLogicServiceStatus defines the observed state of KuberLogicService
type KuberLogicServiceStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
	// namespace that contains service resources
	Namespace string `json:"namespace,omitempty"`
	// date when the namespace and all related resources will be purged
	PurgeDate string `json:"purgeDate,omitempty"`
}

type KuberLogicServiceSpec struct {
	// Type of the cluster
	Type string `json:"type"`
	// Amount of replicas
	// +kubebuilder:validation:Maximum=5
	Replicas int32 `json:"replicas"`
	// Volume size
	VolumeSize string `json:"volumeSize,omitempty"`
	// 2 or 3 digits: 5 or 5.7 or 5.7.31
	// +kubebuilder:validation:Pattern=^\d+[\.\d+]*$
	Version string `json:"version,omitempty"`

	// Resources (requests/limits)
	Limits v1.ResourceList `json:"limits,omitempty"`

	// +kubebuilder:validation:Pattern=[a-z]([-a-z0-9]*[a-z0-9])?
	Domain string `json:"domain,omitempty"`

	// any advanced configuration is supported
	Advanced v11.JSON `json:"advanced,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type == 'Ready')].status",description="The cluster readiness"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type == 'Ready')].reason",description="The cluster status"
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`,description="The cluster type"
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`,description="The number of desired replicas"
// +kubebuilder:printcolumn:name="Volume",type=string,JSONPath=`.spec.volumeSize`,description="Volume size for the cluster"
// +kubebuilder:printcolumn:name="CPU Request",type=string,JSONPath=`.spec.resources.requests.cpu`,description="CPU request"
// +kubebuilder:printcolumn:name="Memory Request",type=string,JSONPath=`.spec.resources.requests.memory`,description="Memory request"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=kls,categories=kuberlogic,scope=Cluster
// +kubebuilder:subresource:status

type KuberLogicService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KuberLogicServiceSpec   `json:"spec,omitempty"`
	Status KuberLogicServiceStatus `json:"status,omitempty"`
}

func (in *KuberLogicService) setConditionStatus(cond string, status bool, msg, reason string) {
	c := metav1.Condition{
		Type:    cond,
		Status:  metav1.ConditionFalse,
		Message: msg,
		Reason:  reason,
	}
	if status {
		c.Status = metav1.ConditionTrue
	}
	meta.SetStatusCondition(&in.Status.Conditions, c)
}

func (in *KuberLogicService) MarkReady(msg string) {
	in.setConditionStatus(readyCondType, true, msg, msg)
}

func (in *KuberLogicService) MarkNotReady(msg string) {
	in.setConditionStatus(readyCondType, false, msg, msg)
}

func (in *KuberLogicService) IsReady() (bool, string) {
	c := meta.FindStatusCondition(in.Status.Conditions, readyCondType)
	if c == nil {
		return false, clusterUnknownStatus
	}
	return c.Status == metav1.ConditionTrue, c.Reason
}

// KuberLogicServiceList contains a list of KuberLogicService
//+kubebuilder:object:root=true
type KuberLogicServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KuberLogicService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KuberLogicService{}, &KuberLogicServiceList{})
}
