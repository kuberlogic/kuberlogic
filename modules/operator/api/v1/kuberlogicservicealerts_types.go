package v1

import (
	"fmt"
	"github.com/kuberlogic/operator/modules/operator/notifications"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KuberLogicAlertSpec struct {
	// Alert name
	// +kubebuilder:validation:Pattern=^.*$
	AlertName string `json:"alertname"`
	// Value
	// +kubebuilder:validation:Pattern=^.*$
	AlertValue string `json:"alertvalue"`
	// Cluster name
	// +kubebuilder:validation:Pattern=^.*$
	Cluster string `json:"cluster"`
	// Pod
	// +kubebuilder:validation:Pattern=^.*$
	Pod string `json:"pod,omitempty"`
	// Silenced set to true suppresses user notification for this alert
	// +kubebuilder:validation:Default=false
	Silenced bool `json:"silenced,omitempty"`
	// Summary contains a descriptive message for this alert
	Summary string `json:"summary,omitempty"`
}

// KuberLogicAlert defines the observed state of KuberLogicAlert
type KuberLogicAlertStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Acknowledged",type="string",JSONPath=".status.conditions[?(@.type == 'Acknowledged')].status",description="Alert has been acknowledged"
// +kubebuilder:printcolumn:name="AlertName",type="string",JSONPath=".spec.alertname",description="Alert name"
// +kubebuilder:printcolumn:name="AlertValue",type="string",JSONPath=".spec.alertvalue",description="Alert value"
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".spec.cluster",description="Cluster name"
// +kubebuilder:printcolumn:name="Pod",type="string",JSONPath=".spec.pod",description="Affected Pod Name"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=kla
// +kubebuilder:subresource:status
type KuberLogicAlert struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KuberLogicAlertSpec   `json:"spec,omitempty"`
	Status KuberLogicAlertStatus `json:"status,omitempty"`
}

// KuberLogicAlertList contains a list of KuberLogicAlert
// +kubebuilder:object:root=true
type KuberLogicAlertList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KuberLogicAlert `json:"items"`
}

const (
	ackCondType      = "Acknowledged"
	notifiedCondType = "NotificationSent"
)

func (kla *KuberLogicAlert) Acknowledge() {
	kla.setConditionStatus(ackCondType, true, "alert has been processed", ackCondType)
}

func (kla *KuberLogicAlert) IsAcknowledged() bool {
	return meta.IsStatusConditionTrue(kla.Status.Conditions, ackCondType)
}

func (kla *KuberLogicAlert) IsSilenced() bool {
	return kla.Spec.Silenced == true
}

func (kla *KuberLogicAlert) NotificationPending() {
	kla.setConditionStatus(notifiedCondType, false, "", notifiedCondType)
}

func (kla *KuberLogicAlert) NotificationSent(addr string) {
	kla.setConditionStatus(notifiedCondType, true, fmt.Sprintf("notification has been sent to %s", addr), notifiedCondType)
}

func (kla *KuberLogicAlert) IsNotificationSent() bool {
	return meta.IsStatusConditionTrue(kla.Status.Conditions, notifiedCondType)
}

// NotifyNew sends a notification through all the required channels (for now email only)
// when a new alert is created.
func (kla *KuberLogicAlert) NotifyNew(addr string) error {
	head := fmt.Sprintf("CRITICAL: SERVICE %s ALERT %s", kla.Spec.Cluster, kla.Spec.AlertName)
	err := notifyEmail(addr, head, kla.Spec.Summary)
	return err
}

// NotifyResolved sends a recovery notification when kla is considered as a resolved (it is implemented with finalizers).
func (kla *KuberLogicAlert) NotifyResolved(addr string) error {
	head := fmt.Sprintf("RESOLVED: SERVICE %s ALERT %s", kla.Spec.Cluster, kla.Spec.AlertName)
	message := fmt.Sprintf("Alert %s is now resolved.", kla.Spec.AlertName)
	err := notifyEmail(addr, head, message)
	return err
}

func (kla *KuberLogicAlert) setConditionStatus(cond string, status bool, msg, reason string) {
	c := metav1.Condition{
		Type:    cond,
		Status:  metav1.ConditionFalse,
		Message: msg,
		Reason:  reason,
	}
	if status {
		c.Status = metav1.ConditionTrue
	}
	meta.SetStatusCondition(&kla.Status.Conditions, c)
}

func notifyEmail(addr, subj, body string) error {
	c, err := notifications.GetNotificationChannel(notifications.EmailChannel)
	if err != nil {
		return err
	}

	opts := map[string]string{
		"to": addr,
	}
	return c.SendNotification(opts, subj, body)
}

func init() {
	SchemeBuilder.Register(&KuberLogicAlert{}, &KuberLogicAlertList{})
}
