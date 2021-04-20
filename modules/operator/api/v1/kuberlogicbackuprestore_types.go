package v1

import (
	"k8s.io/apimachinery/pkg/api/meta"
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
	Conditions []metav1.Condition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Finished",type="string",JSONPath=".status.conditions[?(@.type == 'Finished')].status",description="Restore status"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type == 'Finished')].reason",description="Restore description"
// +kubebuilder:printcolumn:name="Cluster name",type=string,JSONPath=`.spec.name`,description="The cluster name"
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`,description="The backup type"
// +kubebuilder:printcolumn:name="Link",type=string,JSONPath=`.spec.backup`,description="The backup link"
// +kubebuilder:resource:shortName=klr
// +kubebuilder:subresource:status
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

const (
	finishedCondType   = "Finished"
	successfulCondType = "Successful"
)

func (klr *KuberLogicBackupRestore) MarkSuccessfulFinish() {
	klr.setConditionStatus(finishedCondType, true, "", "RestoreFinished")
	klr.setConditionStatus(successfulCondType, true, "", "JobSuccessful")
}

func (klr *KuberLogicBackupRestore) MarkFailed() {
	klr.setConditionStatus(finishedCondType, true, "", "RestoreFinished")
	klr.setConditionStatus(successfulCondType, false, "", "JobFailed")
}

func (klr *KuberLogicBackupRestore) MarkRunning() {
	klr.setConditionStatus(finishedCondType, false, "restore is in progress", "JobIsRunning")
}

func (klr *KuberLogicBackupRestore) setConditionStatus(cond string, status bool, msg, reason string) {
	c := metav1.Condition{
		Type:    cond,
		Status:  metav1.ConditionFalse,
		Message: msg,
		Reason:  reason,
	}
	if status {
		c.Status = metav1.ConditionTrue
	}
	meta.SetStatusCondition(&klr.Status.Conditions, c)
}

func (klr *KuberLogicBackupRestore) IsSuccessful() bool {
	return meta.IsStatusConditionTrue(klr.Status.Conditions, successfulCondType)
}

func init() {
	SchemeBuilder.Register(&KuberLogicBackupRestore{}, &KuberLogicBackupRestoreList{})
}
