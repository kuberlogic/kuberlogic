/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
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
// +kubebuilder:printcolumn:name="Successful",type="string",JSONPath=".status.conditions[?(@.type == 'Successful')].status",description="Restore description"
// +kubebuilder:printcolumn:name="Cluster name",type=string,JSONPath=`.spec.name`,description="The cluster name"
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`,description="The backup type"
// +kubebuilder:printcolumn:name="Link",type=string,JSONPath=`.spec.backup`,description="The backup link"
// +kubebuilder:resource:shortName=klr,categories=kuberlogic
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
	pendingCondType    = "Pending"
	finishedCondType   = "Finished"
	successfulCondType = "Successful"
)

func (klr *KuberLogicBackupRestore) MarkSuccessfulFinish() {
	klr.setConditionStatus(finishedCondType, true, "", "RestoreFinished")
	klr.setConditionStatus(successfulCondType, true, "", "JobSuccessful")
	klr.setConditionStatus(pendingCondType, false, pendingCondType, pendingCondType)
}

func (klr *KuberLogicBackupRestore) MarkFailed() {
	klr.setConditionStatus(finishedCondType, true, "", "RestoreFinished")
	klr.setConditionStatus(successfulCondType, false, "", "JobFailed")
	klr.setConditionStatus(pendingCondType, false, pendingCondType, pendingCondType)
}

func (klr *KuberLogicBackupRestore) MarkRunning() {
	klr.setConditionStatus(finishedCondType, false, "restore is in progress", "JobIsRunning")
	klr.setConditionStatus(pendingCondType, false, finishedCondType, finishedCondType)
}

func (klr *KuberLogicBackupRestore) MarkPending() {
	klr.setConditionStatus(pendingCondType, true, pendingCondType, pendingCondType)
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

// returns completion description and time
func (klr *KuberLogicBackupRestore) GetCompletionStatus() (string, *time.Time) {
	c := meta.FindStatusCondition(klr.Status.Conditions, finishedCondType)
	if c == nil {
		return RestoreUnknownStatus, nil
	}
	if c.Status == metav1.ConditionFalse {
		return RestoreRunningStatus, nil
	}

	successCond := meta.FindStatusCondition(klr.Status.Conditions, successfulCondType)
	if successCond == nil {
		return RestoreUnknownStatus, nil
	}
	compTime := successCond.LastTransitionTime.Time.UTC()
	switch successCond.Status {
	case metav1.ConditionTrue:
		return RestoreSuccessStatus, &compTime
	default:
		return RestoreFailedStatus, &compTime
	}
}

func init() {
	SchemeBuilder.Register(&KuberLogicBackupRestore{}, &KuberLogicBackupRestoreList{})
}
