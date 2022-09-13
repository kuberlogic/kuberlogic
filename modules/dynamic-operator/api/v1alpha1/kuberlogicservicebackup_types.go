/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	klbSuccessfulCondType = "Successful"
	klbFailedCondType     = "Failed"
	klbRequestedCondType  = "Requested"
)

// KuberlogicServiceBackupSpec defines the desired state of KuberlogicServiceBackup
type KuberlogicServiceBackupSpec struct {
	KuberlogicServiceName string `json:"kuberlogicServiceName"`
}

// KuberlogicServiceBackupStatus defines the observed state of KuberlogicServiceBackup
type KuberlogicServiceBackupStatus struct {
	Conditions      []metav1.Condition `json:"conditions"`
	Phase           string             `json:"phase,omitempty"`
	BackupReference string             `json:"backupReference,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=klb,categories=kuberlogic,scope=Cluster
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase",description="Backup status"

// KuberlogicServiceBackup is the Schema for the kuberlogicservicebackups API
type KuberlogicServiceBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KuberlogicServiceBackupSpec   `json:"spec,omitempty"`
	Status KuberlogicServiceBackupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KuberlogicServiceBackupList contains a list of KuberlogicServiceBackup
type KuberlogicServiceBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KuberlogicServiceBackup `json:"items"`
}

func (in *KuberlogicServiceBackup) IsFailed() bool {
	return in.Status.Phase == klbFailedCondType
}

func (in *KuberlogicServiceBackup) IsSuccessful() bool {
	return in.Status.Phase == klbSuccessfulCondType
}

func (in *KuberlogicServiceBackup) IsRequested() bool {
	return in.Status.Phase == klbRequestedCondType
}

func (in *KuberlogicServiceBackup) MarkFailed(reason string) {
	in.Status.Phase = klbFailedCondType
	in.setConditionStatus(klbFailedCondType, true, reason, klbFailedCondType)
	in.setConditionStatus(klbSuccessfulCondType, false, reason, klbSuccessfulCondType)
}

func (in *KuberlogicServiceBackup) MarkSuccessful() {
	in.Status.Phase = klbSuccessfulCondType
	in.setConditionStatus(klbSuccessfulCondType, true, "", klbSuccessfulCondType)
	in.setConditionStatus(klbFailedCondType, false, "", klbFailedCondType)
}

func (in *KuberlogicServiceBackup) MarkRequested() {
	in.Status.Phase = klbRequestedCondType
	in.setConditionStatus(klbRequestedCondType, true, "", klbRequestedCondType)
}

func (in *KuberlogicServiceBackup) setConditionStatus(cond string, status bool, msg, reason string) {
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

func init() {
	SchemeBuilder.Register(&KuberlogicServiceBackup{}, &KuberlogicServiceBackupList{})
}
