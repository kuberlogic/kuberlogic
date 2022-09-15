/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	KlbSuccessfulCondType = "Successful"
	KlbFailedCondType     = "Failed"
	KlbRequestedCondType  = "Requested"
)

// KuberlogicServiceBackupSpec defines the desired state of KuberlogicServiceBackup
type KuberlogicServiceBackupSpec struct {
	KuberlogicServiceName string `json:"kuberlogicServiceName"`
}

// KuberlogicServiceBackupStatus defines the observed state of KuberlogicServiceBackup
type KuberlogicServiceBackupStatus struct {
	Conditions     []metav1.Condition `json:"conditions"`
	Phase          string             `json:"phase,omitempty"`
	FailedAttempts int                `json:"FailedAttempts,omitempty"`
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
	return in.Status.Phase == KlbFailedCondType
}

func (in *KuberlogicServiceBackup) IsSuccessful() bool {
	return in.Status.Phase == KlbSuccessfulCondType
}

func (in *KuberlogicServiceBackup) IsPending() bool {
	return !(in.IsFailed() || in.IsSuccessful() || in.IsRequested())
}

func (in *KuberlogicServiceBackup) IsRequested() bool {
	return in.Status.Phase == KlbRequestedCondType
}

func (in *KuberlogicServiceBackup) MarkFailed(reason string) {
	in.Status.Phase = KlbFailedCondType
	in.setConditionStatus(KlbFailedCondType, true, reason, KlbFailedCondType)
	in.setConditionStatus(KlbSuccessfulCondType, false, reason, KlbSuccessfulCondType)
}

func (in *KuberlogicServiceBackup) MarkSuccessful() {
	in.Status.Phase = KlbSuccessfulCondType
	in.setConditionStatus(KlbSuccessfulCondType, true, "", KlbSuccessfulCondType)
	in.setConditionStatus(KlbFailedCondType, false, "", KlbFailedCondType)
}

func (in *KuberlogicServiceBackup) MarkRequested() {
	in.Status.Phase = KlbRequestedCondType
	in.setConditionStatus(KlbRequestedCondType, true, "", KlbRequestedCondType)
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

func (in *KuberlogicServiceBackup) IncreaseFailedAttemptCount() {
	in.Status.FailedAttempts += 1
}

func init() {
	SchemeBuilder.Register(&KuberlogicServiceBackup{}, &KuberlogicServiceBackupList{})
}
