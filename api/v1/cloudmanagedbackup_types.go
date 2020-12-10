package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CloudManagedBackupSpec struct {
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

// CloudManagedBackupStatus defines the observed state of CloudManagedBackup
type CloudManagedBackupStatus struct {
	Status string `json:"status"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status",description="The backup status"
// +kubebuilder:printcolumn:name="Cluster name",type=string,JSONPath=`.spec.name`,description="The cluster name"
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`,description="The backup type"
// +kubebuilder:printcolumn:name="Schedule",type=string,JSONPath=`.spec.schedule`,description="The backup schedule"
// +kubebuilder:resource:shortName=clb
type CloudManagedBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudManagedBackupSpec   `json:"spec,omitempty"`
	Status CloudManagedBackupStatus `json:"status,omitempty"`
}

// CloudManagedBackupList contains a list of CloudManagedBackup
// +kubebuilder:object:root=true
type CloudManagedBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudManagedBackup `json:"items"`
}

func (cm *CloudManagedBackup) IsEqual(newStatus string) bool {
	return cm.Status.Status == newStatus
}

func (cm *CloudManagedBackup) SetStatus(newStatus string) {
	cm.Status.Status = newStatus
}

func (cm *CloudManagedBackup) GetStatus() string {
	return cm.Status.Status
}

func init() {
	SchemeBuilder.Register(&CloudManagedBackup{}, &CloudManagedBackupList{})
}
