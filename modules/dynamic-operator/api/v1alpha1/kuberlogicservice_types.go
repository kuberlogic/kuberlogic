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
	readyCondType          = "Ready"
	clusterUnknownStatus   = "Unknown"
	pausedCondType         = "Paused"
	backupRunningCondType  = "BackupRunning"
	restoreRunningCondType = "RestoreRunning"
)

// KuberLogicServiceStatus defines the observed state of KuberLogicService
type KuberLogicServiceStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
	// namespace that contains service resources
	Namespace string `json:"namespace,omitempty"`
	// date when the namespace and all related resources will be purged
	PurgeDate string `json:"purgeDate,omitempty"`

	AccessEndpoint string `json:"access,omitempty"`

	// a service is about to be restored or restore is in progress
	RestoreRequested bool `json:"restoreRequested,omitempty"`
	// a service is ready for restore process
	ReadyForRestore bool `json:"readyForRestore,omitempty"`
}

type KuberLogicServiceSpec struct {
	// Type of the cluster
	Type string `json:"type"`
	// Amount of replicas
	// +kubebuilder:validation:Maximum=5
	Replicas int32 `json:"replicas,omitempty"`
	// Volume size
	VolumeSize string `json:"volumeSize,omitempty"`
	// 2 or 3 digits: 5 or 5.7 or 5.7.31
	// +kubebuilder:validation:Pattern=^\d+[\.\d+]*$
	Version string `json:"version,omitempty"`

	// Resources (requests/limits)
	Limits v1.ResourceList `json:"limits,omitempty"`

	// +kubebuilder:validation:Pattern=[a-z]([-a-z0-9]*[a-z0-9])?
	Domain     string `json:"domain,omitempty"`
	TLSEnabled bool   `json:"TLSEnabled,omitempty"`

	// any advanced configuration is supported
	Advanced v11.JSON `json:"advanced,omitempty"`

	// Paused field allows to stop all service related containers
	// +kubebuilder:default=false
	Paused bool `json:"paused,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type == 'Ready')].status",description="The cluster readiness"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type == 'Ready')].reason",description="The cluster status"
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`,description="The cluster type"
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`,description="The number of desired replicas"
// +kubebuilder:printcolumn:name="Volume",type=string,JSONPath=`.spec.volumeSize`,description="Volume size for the cluster"
// +kubebuilder:printcolumn:name="CPU Limits",type=string,JSONPath=`.spec.limits.cpu`,description="CPU limits"
// +kubebuilder:printcolumn:name="Memory Limits",type=string,JSONPath=`.spec.limits.memory`,description="Memory limits"
// +kubebuilder:printcolumn:name="Endpoint",type=string,JSONPath=`.status.access`,description="Access endpoint"
// +kubebuilder:printcolume:name="Paused",type=bool,JSONPath=`.spec.paused`,description="Service is paused"
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

func (in *KuberLogicService) Paused() bool {
	return in.Spec.Paused
}

func (in *KuberLogicService) TLSEnabled() bool {
	return in.Spec.TLSEnabled
}

func (in *KuberLogicService) GetHost() string {
	var host string
	if in.Spec.Domain != "" {
		host = in.GetName() + "." + in.Spec.Domain
	}
	return host
}

func (in *KuberLogicService) SetAccessEndpoint() {
	host := in.GetHost()
	if host == "" {
		return
	}
	proto := "http://"
	if in.Spec.TLSEnabled {
		proto = "https://"
	}
	in.Status.AccessEndpoint = proto + host
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

func (in *KuberLogicService) MarkPaused() {
	in.setConditionStatus(pausedCondType, true, pausedCondType, pausedCondType)
}

func (in *KuberLogicService) MarkResumed() {
	in.setConditionStatus(pausedCondType, false, pausedCondType, pausedCondType)
}

// PauseRequested indicates that a kls pause is requested
func (in *KuberLogicService) PauseRequested() bool {
	c := meta.FindStatusCondition(in.Status.Conditions, pausedCondType)
	return in.Spec.Paused && (c == nil || c.Status == metav1.ConditionFalse)
}

// Resumed indicates that a kls was paused and now is resumed by setting in.Spec.Paused false
func (in *KuberLogicService) Resumed() bool {
	c := meta.FindStatusCondition(in.Status.Conditions, pausedCondType)
	if c == nil {
		return false
	}
	return c.Status == metav1.ConditionTrue && !in.Spec.Paused
}

func (in *KuberLogicService) RestoreRunning() (bool, string) {
	c := meta.FindStatusCondition(in.Status.Conditions, restoreRunningCondType)
	if c == nil {
		return false, ""
	}
	return c.Status == metav1.ConditionTrue, c.Message
}

func (in *KuberLogicService) SetRestoreStatus(klr *KuberlogicServiceRestore) {
	in.setConditionStatus(restoreRunningCondType, !(klr.IsFailed() || klr.IsSuccessful()), klr.Name, restoreRunningCondType)
}

func (in *KuberLogicService) BackupRunning() (bool, string) {
	c := meta.FindStatusCondition(in.Status.Conditions, backupRunningCondType)
	if c == nil {
		return false, ""
	}
	return c.Status == metav1.ConditionTrue, c.Message
}

func (in *KuberLogicService) SetBackupStatus(klb *KuberlogicServiceBackup) {
	in.setConditionStatus(backupRunningCondType, !(klb.IsFailed() || klb.IsSuccessful()), klb.Name, backupRunningCondType)
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
