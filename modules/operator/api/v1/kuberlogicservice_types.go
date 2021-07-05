package v1

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
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
	// 2 or 3 digits: 5 or 5.7 or 5.7.31
	// +kubebuilder:validation:Pattern=^\d+[\.\d+]*$
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
	Conditions []metav1.Condition `json:"conditions"`
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
// +kubebuilder:resource:shortName=kls,categories=kuberlogic
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

const (
	readyCondType             = "Ready"
	backupInProgressCondType  = "BackupInProgress"
	restoreInProgressCondType = "RestoreInProgress"
)

func (kls *KuberLogicService) MarkReady(msg string) {
	kls.setConditionStatus(readyCondType, true, msg, msg)
}

func (kls *KuberLogicService) MarkNotReady(msg string) {
	kls.setConditionStatus(readyCondType, false, msg, msg)
}

func (kls *KuberLogicService) ReconciliationAllowed() bool {
	return meta.IsStatusConditionTrue(kls.Status.Conditions, readyCondType) ||
		meta.IsStatusConditionFalse(kls.Status.Conditions, backupInProgressCondType) ||
		meta.IsStatusConditionFalse(kls.Status.Conditions, restoreInProgressCondType)
}

func (kls *KuberLogicService) IsReady() (bool, string) {
	c := meta.FindStatusCondition(kls.Status.Conditions, readyCondType)
	if c == nil {
		return false, ClusterUnknownStatus
	}
	return c.Status == metav1.ConditionTrue, c.Reason
}

func (kls *KuberLogicService) BackupRunning(name string) {
	kls.setConditionStatus(backupInProgressCondType, true, name+" backup job is running", "BackupRunning")
}

func (kls *KuberLogicService) BackupFinished() {
	kls.setConditionStatus(backupInProgressCondType, false, "", "NoBackupRunning")
}

func (kls *KuberLogicService) RestoreStarted(name string) {
	kls.setConditionStatus(restoreInProgressCondType, true, name+" restore job is running", "RestoreRunning")
}

func (kls *KuberLogicService) RestoreFinished() {
	kls.setConditionStatus(restoreInProgressCondType, false, "", "NoRestoreRunning")
}

func (kls *KuberLogicService) SetAlertEmail(e string) error {
	if e == "" {
		return fmt.Errorf("email can't be empty")
	}
	m, err := regexp.MatchString(".*@.*", e)
	if !m {
		return fmt.Errorf("incorrect email address")
	}
	if err != nil {
		return err
	}

	kls.SetAnnotations(map[string]string{
		alertEmailAnnotation: e,
	})
	return nil
}

func (kls *KuberLogicService) GetAlertEmail() string {
	a := kls.GetAnnotations()
	email, found := a[alertEmailAnnotation]
	if found {
		return email
	}
	return ""
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

func (kls *KuberLogicService) setConditionStatus(cond string, status bool, msg, reason string) {
	c := metav1.Condition{
		Type:    cond,
		Status:  metav1.ConditionFalse,
		Message: msg,
		Reason:  reason,
	}
	if status {
		c.Status = metav1.ConditionTrue
	}
	meta.SetStatusCondition(&kls.Status.Conditions, c)
}

func init() {
	SchemeBuilder.Register(&KuberLogicService{}, &KuberLogicServiceList{})
}
