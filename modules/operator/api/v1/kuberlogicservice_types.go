package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KuberLogicServiceSpec struct {
	// Type of the cluster
	// +kubebuilder:validation:Enum=postgresql;mysql;redis
	Type string `json:"type"`
	// Amount of replicas
	// +kubebuilder:validation:Maximum=5
	Replicas int32 `json:"replicas"`
	// Resources (requests/limits)
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
	// Volume size
	VolumeSize string `json:"volumeSize,omitempty"`
	// 2 or 3 digits: 5.7 or 5.7.31
	// +kubebuilder:validation:Pattern=^\d+\.\d+[\.\d+]*$
	Version           string            `json:"version,omitempty"`
	AdvancedConf      map[string]string `json:"advancedConf,omitempty"`
	MaintenanceWindow `json:"maintenanceWindow,omitempty"`
}

type MaintenanceWindow struct {
	// start hour, UTC zone is assumed
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=23
	// +kubebuilder:validation:Type=integer
	StartHour int `json:"start,omitempty"`
	// day of the week
	// +kubebuilder:validation:Enum=Monday;Tuesday;Wednesday;Thursday;Friday;Saturday;Sunday
	Weekday string `json:"weekday,omitempty"`
	// window duration in hours
	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Default=4
	DurationHours int `json:"duration,omitempty"`
}

// KuberLogicServiceStatus defines the observed state of KuberLogicService
type KuberLogicServiceStatus struct {
	Status string `json:"status"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status",description="The cluster status"
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`,description="The cluster type"
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`,description="The number of desired replicas"
// +kubebuilder:printcolumn:name="Volume",type=string,JSONPath=`.spec.volumeSize`,description="Volume size for the cluster"
// +kubebuilder:printcolumn:name="CPU Request",type=string,JSONPath=`.spec.resources.requests.cpu`,description="CPU request"
// +kubebuilder:printcolumn:name="Memory Request",type=string,JSONPath=`.spec.resources.requests.memory`,description="Memory request"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=kls
// +kubebuilder:subresource:status
type KuberLogicService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KuberLogicServiceSpec   `json:"spec,omitempty"`
	Status KuberLogicServiceStatus `json:"status,omitempty"`
}

// KuberLogicServiceList contains a list of KuberLogicService
// +kubebuilder:object:root=true
type KuberLogicServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KuberLogicService `json:"items"`
}

func (kls *KuberLogicService) IsEqual(newStatus string) bool {
	return kls.Status.Status == newStatus
}

func (kls *KuberLogicService) SetStatus(newStatus string) {
	kls.Status.Status = newStatus
}

func (kls *KuberLogicService) GetStatus() string {
	return kls.Status.Status
}

func (kls *KuberLogicService) UpdatesAllowed() bool {
	return kls.Status.Status == ClusterOkStatus ||
		kls.Status.Status == ClusterFailedStatus
}

// TODO: Figure out workaround in https://github.com/kubernetes-sigs/kubebuilder/issues/1501, not it's a blocker
// for implementation default values based on webhook (https://book.kubebuilder.io/cronjob-tutorial/webhook-implementation.html)
func (kls *KuberLogicService) InitDefaults(defaults Defaults) bool {
	dirty := false
	if kls.Spec.Resources.Requests == nil {
		kls.Spec.Resources.Requests = defaults.Resources.Requests
		dirty = true
	}
	if kls.Spec.Resources.Limits == nil {
		kls.Spec.Resources.Limits = defaults.Resources.Limits
		dirty = true
	}
	if kls.Spec.VolumeSize == "" {
		kls.Spec.VolumeSize = defaults.VolumeSize
		dirty = true
	}
	if kls.Spec.Version == "" {
		kls.Spec.Version = defaults.Version
		dirty = true
	}

	return dirty
}

func init() {
	SchemeBuilder.Register(&KuberLogicService{}, &KuberLogicServiceList{})
}
