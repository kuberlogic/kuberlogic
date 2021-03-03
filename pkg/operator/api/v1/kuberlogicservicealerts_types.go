package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KuberLogicAlertSpec struct {
	// Alert name
	// +kubebuilder:validation:Pattern=^.*$
	AlertName string `json:"alertname"`
	// Value
	// +kubebuilder:validation:Pattern=^.*$
	AlertValue string `json:"alertvalue"`
	// Cluster name
	// +kubebuilder:validation:Pattern=^.*$
	Cluster string `json:"cluster"`
	// Pod
	// +kubebuilder:validation:Pattern=^.*$
	Pod string `json:"pod"`
}

// KuberLogicAlert defines the observed state of KuberLogicAlert
type KuberLogicAlertStatus struct {
	Status string `json:"status"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status",description="The cluster status"
// +kubebuilder:printcolumn:name="AlertName",type="string",JSONPath=".spec.alertname",description="Alert name"
// +kubebuilder:printcolumn:name="AlertValue",type="string",JSONPath=".spec.alertvalue",description="Alert value"
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".spec.cluster",description="Cluster name"
// +kubebuilder:printcolumn:name="Pod",type="string",JSONPath=".spec.pod",description="Affected Pod Name"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=kla
type KuberLogicAlert struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KuberLogicAlertSpec   `json:"spec,omitempty"`
	Status KuberLogicAlertStatus `json:"status,omitempty"`
}

// KuberLogicAlertList contains a list of KuberLogicAlert
// +kubebuilder:object:root=true
type KuberLogicAlertList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KuberLogicAlert `json:"items"`
}

func (kla *KuberLogicAlert) IsEqual(newStatus string) bool {
	return kla.Status.Status == newStatus
}

func (kla *KuberLogicAlert) SetStatus(newStatus string) {
	kla.Status.Status = newStatus
}

func (kla *KuberLogicAlert) GetStatus() string {
	return kla.Status.Status
}

func init() {
	SchemeBuilder.Register(&KuberLogicAlert{}, &KuberLogicAlertList{})
}
