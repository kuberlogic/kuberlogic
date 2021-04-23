package v1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KuberLogicBackupScheduleSpec struct {
	// Type of the backup storage
	// +kubebuilder:validation:Enum=s3
	Type string `json:"type"`
	// Cluster name
	// TODO: need to implement validation based on webhook - https://book.kubebuilder.io/cronjob-tutorial/webhook-implementation.html
	// it gives the ability to check cluster name is exists or not
	ClusterName string `json:"name"`
	// credentials for storage type
	SecretName string `json:"secret"`
	// schedule for the backup
	Schedule string `json:"schedule"`
	// what database need to backup
	Database string `json:"database,omitempty"`
}

// KuberLogicBackupScheduleStatus defines the observed state of KuberLogicBackupSchedule
type KuberLogicBackupScheduleStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Last backup status",type=string,JSONPath=".status.conditions[?(@.type == 'LastBackupSuccessful')].reason",description="Current backup status"
// +kubebuilder:printcolumn:name="Cluster name",type=string,JSONPath=`.spec.name`,description="The cluster name"
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`,description="The backup type"
// +kubebuilder:printcolumn:name="Schedule",type=string,JSONPath=`.spec.schedule`,description="The backup schedule"
// +kubebuilder:resource:shortName=klb
// +kubebuilder:subresource:status
type KuberLogicBackupSchedule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KuberLogicBackupScheduleSpec   `json:"spec,omitempty"`
	Status KuberLogicBackupScheduleStatus `json:"status,omitempty"`
}

// KuberLogicBackupScheduleList contains a list of KuberLogicBackupSchedule
// +kubebuilder:object:root=true
type KuberLogicBackupScheduleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KuberLogicBackupSchedule `json:"items"`
}

const (
	lastJobStateCond = "LastBackupSuccessful"
	runningCondType  = "Running"
)

func (klb *KuberLogicBackupSchedule) MarkFailed(r string) {
	klb.setConditionStatus(lastJobStateCond, false, r, BackupFailedStatus)
}

func (klb *KuberLogicBackupSchedule) MarkUnknown(m string) {
	klb.setConditionStatus(lastJobStateCond, false, m, BackupUnknownStatus)
}

func (klb *KuberLogicBackupSchedule) MarkSuccessful(m string) {
	klb.setConditionStatus(lastJobStateCond, true, m, BackupSuccessStatus)
}

func (klb *KuberLogicBackupSchedule) MarkRunning(j string) {
	klb.setConditionStatus(runningCondType, true, j+" backup job is running", "BackupJobRunning")
}

func (klb *KuberLogicBackupSchedule) MarkNotRunning() {
	klb.setConditionStatus(runningCondType, false, "", "NoJobRunning")
}

func (klb *KuberLogicBackupSchedule) IsRunning() bool {
	return meta.IsStatusConditionTrue(klb.Status.Conditions, runningCondType)
}

func (klb *KuberLogicBackupSchedule) IsSuccessful() bool {
	return meta.IsStatusConditionTrue(klb.Status.Conditions, lastJobStateCond)
}

func (klb *KuberLogicBackupSchedule) setConditionStatus(cond string, status bool, msg, reason string) {
	c := metav1.Condition{
		Type:    cond,
		Status:  metav1.ConditionFalse,
		Message: msg,
		Reason:  reason,
	}
	if status {
		c.Status = metav1.ConditionTrue
	}
	meta.SetStatusCondition(&klb.Status.Conditions, c)
}

func init() {
	SchemeBuilder.Register(&KuberLogicBackupSchedule{}, &KuberLogicBackupScheduleList{})
}
