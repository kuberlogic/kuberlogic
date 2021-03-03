package v1

import (
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
	Status string `json:"status"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status",description="The backup status"
// +kubebuilder:printcolumn:name="Cluster name",type=string,JSONPath=`.spec.name`,description="The cluster name"
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`,description="The backup type"
// +kubebuilder:printcolumn:name="Schedule",type=string,JSONPath=`.spec.schedule`,description="The backup schedule"
// +kubebuilder:resource:shortName=klb
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

func (klb *KuberLogicBackupSchedule) IsEqual(newStatus string) bool {
	return klb.Status.Status == newStatus
}

func (klb *KuberLogicBackupSchedule) SetStatus(newStatus string) {
	klb.Status.Status = newStatus
}

func (klb *KuberLogicBackupSchedule) GetStatus() string {
	return klb.Status.Status
}

func init() {
	SchemeBuilder.Register(&KuberLogicBackupSchedule{}, &KuberLogicBackupScheduleList{})
}
