package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CloudManagedRestoreSpec struct {
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

// CloudManagedRestoreStatus defines the observed state of CloudManagedRestore
type CloudManagedRestoreStatus struct {
	Status string `json:"status"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status",description="The backup status"
// +kubebuilder:printcolumn:name="Cluster name",type=string,JSONPath=`.spec.name`,description="The cluster name"
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`,description="The backup type"
// +kubebuilder:printcolumn:name="Link",type=string,JSONPath=`.spec.backup`,description="The backup link"
// +kubebuilder:resource:shortName=clr
type CloudManagedRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudManagedRestoreSpec   `json:"spec,omitempty"`
	Status CloudManagedRestoreStatus `json:"status,omitempty"`
}

// CloudManagedRestoreList contains a list of CloudManagedRestore
// +kubebuilder:object:root=true
type CloudManagedRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudManagedRestore `json:"items"`
}

func (cm *CloudManagedRestore) IsEqual(newStatus string) bool {
	return cm.Status.Status == newStatus
}

func (cm *CloudManagedRestore) SetStatus(newStatus string) {
	cm.Status.Status = newStatus
}

func (cm *CloudManagedRestore) GetStatus() string {
	return cm.Status.Status
}

func init() {
	SchemeBuilder.Register(&CloudManagedRestore{}, &CloudManagedRestoreList{})
}
