/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	klrFailedCondType     = "Failed"
	klrSuccessfulCondType = "Successful"
	klrRequestedCondType  = "Requested"
)

// KuberlogicServiceRestoreSpec defines the desired state of KuberlogicServiceRestore
type KuberlogicServiceRestoreSpec struct {
	KuberlogicServiceBackup string `json:"kuberlogicServiceBackup"`
}

// KuberlogicServiceRestoreStatus defines the observed state of KuberlogicServiceRestore
type KuberlogicServiceRestoreStatus struct {
	RestoreReference string             `json:"restoreReference,omitempty"`
	Conditions       []metav1.Condition `json:"conditions"`
	Phase            string             `json:"phase,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=klr,categories=kuberlogic,scope=Cluster
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase",description="Restore status"

// KuberlogicServiceRestore is the Schema for the kuberlogicservicerestores API
type KuberlogicServiceRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KuberlogicServiceRestoreSpec   `json:"spec,omitempty"`
	Status KuberlogicServiceRestoreStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KuberlogicServiceRestoreList contains a list of KuberlogicServiceRestore
type KuberlogicServiceRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KuberlogicServiceRestore `json:"items"`
}

func (in *KuberlogicServiceRestore) setConditionStatus(cond string, status bool, msg, reason string) {
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

func (in *KuberlogicServiceRestore) MarkFailed(reason string) {
	in.Status.Phase = klrFailedCondType
	in.setConditionStatus(klrFailedCondType, true, reason, klrFailedCondType)
	in.setConditionStatus(klrSuccessfulCondType, false, "", klrSuccessfulCondType)
}

func (in *KuberlogicServiceRestore) MarkSuccessful() {
	in.Status.Phase = klrSuccessfulCondType
	in.setConditionStatus(klrFailedCondType, false, "", klrFailedCondType)
	in.setConditionStatus(klrSuccessfulCondType, true, "", klrSuccessfulCondType)
}

func (in *KuberlogicServiceRestore) MarkRequested() {
	in.Status.Phase = klrRequestedCondType
	in.setConditionStatus(klrRequestedCondType, true, "", klrRequestedCondType)
}

func (in *KuberlogicServiceRestore) IsFailed() bool {
	return in.Status.Phase == klrFailedCondType
}

func (in *KuberlogicServiceRestore) IsSuccessful() bool {
	return in.Status.Phase == klrSuccessfulCondType
}

func (in *KuberlogicServiceRestore) IsRequested() bool {
	return in.Status.Phase == klrRequestedCondType
}

func init() {
	SchemeBuilder.Register(&KuberlogicServiceRestore{}, &KuberlogicServiceRestoreList{})
}
