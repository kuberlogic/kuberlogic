/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KuberlogicServiceBackupScheduleSpec defines the desired state of KuberlogicServiceBackupSchedule
type KuberlogicServiceBackupScheduleSpec struct {
	KuberlogicServiceName string `json:"kuberlogicServiceName"`
	Schedule              string `json:"schedule,omitempty"`
}

// KuberlogicServiceBackupScheduleStatus defines the observed state of KuberlogicServiceBackupSchedule
type KuberlogicServiceBackupScheduleStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=klbs,categories=kuberlogic,scope=Namespaced

// KuberlogicServiceBackupSchedule is the Schema for the kuberlogicservicebackupschedules API
type KuberlogicServiceBackupSchedule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KuberlogicServiceBackupScheduleSpec   `json:"spec,omitempty"`
	Status KuberlogicServiceBackupScheduleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KuberlogicServiceBackupScheduleList contains a list of KuberlogicServiceBackupSchedule
type KuberlogicServiceBackupScheduleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KuberlogicServiceBackupSchedule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KuberlogicServiceBackupSchedule{}, &KuberlogicServiceBackupScheduleList{})
}
