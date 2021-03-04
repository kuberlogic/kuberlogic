package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KuberLogicBackupRestoreSpec struct {
	// Type of the backup storage
	// +kubebuilder:validation:Enum=s3
	Type string `json:"type"`
	// Cluster name
	// TODO: need to implement validation based on webhook - https://book.kubebuilder.io/cronjob-tutorial/webhook-implementation.html
	// it gives the ability to check cluster name is exists or not
	ClusterName string `json:"name"`
	// credentials for storage type
	SecretName string `json:"secret"`
	// link to backup on the storage
	Backup string `json:"backup"`
	// what database need to be restored
	Database string `json:"database"`
}

// KuberLogicBackupRestoreStatus defines the observed state of KuberLogicBackupRestore
type KuberLogicBackupRestoreStatus struct {
	Status string `json:"status"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status",description="The backup status"
// +kubebuilder:printcolumn:name="Cluster name",type=string,JSONPath=`.spec.name`,description="The cluster name"
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`,description="The backup type"
// +kubebuilder:printcolumn:name="Link",type=string,JSONPath=`.spec.backup`,description="The backup link"
// +kubebuilder:resource:shortName=klr
type KuberLogicBackupRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KuberLogicBackupRestoreSpec   `json:"spec,omitempty"`
	Status KuberLogicBackupRestoreStatus `json:"status,omitempty"`
}

// KuberLogicBackupRestoreList contains a list of KuberLogicBackupRestore
// +kubebuilder:object:root=true
type KuberLogicBackupRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KuberLogicBackupRestore `json:"items"`
}

func (klr *KuberLogicBackupRestore) IsEqual(newStatus string) bool {
	return klr.Status.Status == newStatus
}

func (klr *KuberLogicBackupRestore) SetStatus(newStatus string) {
	klr.Status.Status = newStatus
}

func (klr *KuberLogicBackupRestore) GetStatus() string {
	return klr.Status.Status
}

func init() {
	SchemeBuilder.Register(&KuberLogicBackupRestore{}, &KuberLogicBackupRestoreList{})
}
