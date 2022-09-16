/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package v1alpha1

import (
	"time"

	v1 "k8s.io/api/core/v1"
	v11 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// CredsUpdateSecretName is a corev1.Secret name that is created when a credentials update operation is requested via KL apiserver
	CredsUpdateSecretName = "credential-request"

	configFailedCondType       = "ConfigurationError"
	provisioningFailedCondType = "ProvisioningError"
	clusterUnknownStatus       = "Unknown"
	pausedCondType             = "Paused"
	backupRunningCondType      = "BackupRunning"
	restoreRunningCondType     = "RestoreRunning"
	ReadyCondType              = "Ready"
)

// KuberLogicServiceStatus defines the observed state of KuberLogicService
type KuberLogicServiceStatus struct {
	Phase      string             `json:"phase,omitempty"`
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
	// 2 or 3 digits: 5 or 5.7 or 5.7.31
	// +kubebuilder:validation:Pattern=^\d+[\.\d+]*$
	Version string `json:"version,omitempty"`

	// Resources (requests/limits)
	Limits v1.ResourceList `json:"limits,omitempty"`

	// +kubebuilder:validation:Pattern=[a-z]([-a-z0-9]*[a-z0-9])?
	Domain   string `json:"domain,omitempty"`
	Insecure bool   `json:"insecure,omitempty"`

	// any advanced configuration is supported
	Advanced v11.JSON `json:"advanced,omitempty"`

	// Paused field allows to stop all service related containers
	// +kubebuilder:default=false
	Paused bool `json:"paused,omitempty"`

	BackupSchedule string `json:"backupSchedule,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase",description="Service status"
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`,description="The cluster type"
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`,description="The number of desired replicas"
// +kubebuilder:printcolumn:name="CPU Limits",type=string,JSONPath=`.spec.limits.cpu`,description="CPU limits"
// +kubebuilder:printcolumn:name="Memory Limits",type=string,JSONPath=`.spec.limits.memory`,description="Memory limits"
// +kubebuilder:printcolumn:name="Storage Limits",type=string,JSONPath=`.spec.limits.storage`,description="Storage limits"
// +kubebuilder:printcolumn:name="Endpoint",type=string,JSONPath=`.status.access`,description="Access endpoint"
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

func (in *KuberLogicService) Insecure() bool {
	return in.Spec.Insecure
}

func (in *KuberLogicService) GetHost() string {
	return in.Spec.Domain
}

func (in *KuberLogicService) SetAccessEndpoint() {
	host := in.GetHost()
	if host == "" {
		return
	}
	proto := "https://"
	if in.Spec.Insecure {
		proto = "http://"
	}
	in.Status.AccessEndpoint = proto + host
}

func (in *KuberLogicService) MarkReady(msg string) {
	in.Status.Phase = ReadyCondType

	in.setConditionStatus(configFailedCondType, false, "", configFailedCondType)
	in.setConditionStatus(provisioningFailedCondType, false, "", provisioningFailedCondType)
	in.setConditionStatus(ReadyCondType, true, msg, msg)
}

func (in *KuberLogicService) MarkNotReady(msg string) {
	in.Status.Phase = "NotReady"
	in.setConditionStatus(configFailedCondType, false, "", configFailedCondType)
	in.setConditionStatus(provisioningFailedCondType, false, "", provisioningFailedCondType)
	in.setConditionStatus(ReadyCondType, false, msg, msg)
}

// IsReady returns
// * true if ready
// * string containing current status
// * time of last readiness status transition
func (in *KuberLogicService) IsReady() (bool, string, *time.Time) {
	c := meta.FindStatusCondition(in.Status.Conditions, ReadyCondType)
	if c == nil {
		return false, clusterUnknownStatus, nil
	}
	return c.Status == metav1.ConditionTrue, c.Reason, &c.LastTransitionTime.Time
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
	if klr == nil {
		in.setConditionStatus(restoreRunningCondType, false, "restore is nil", restoreRunningCondType)
		return
	}
	in.Status.Phase = "Restoring"
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
	if klb == nil {
		in.setConditionStatus(backupRunningCondType, false, "backup is nil", backupRunningCondType)
		return
	}
	in.Status.Phase = "Backing Up"
	in.setConditionStatus(backupRunningCondType, !(klb.IsFailed() || klb.IsSuccessful()), klb.Name, backupRunningCondType)
}

func (in *KuberLogicService) ConfigurationFailed(s string) {
	in.Status.Phase = configFailedCondType
	in.setConditionStatus(configFailedCondType, true, s, configFailedCondType)
}

func (in *KuberLogicService) ClusterSyncFailed(s string) {
	in.Status.Phase = provisioningFailedCondType
	in.setConditionStatus(provisioningFailedCondType, true, s, provisioningFailedCondType)
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
