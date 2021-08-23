package v1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KuberLogicTenantSpec struct {
	OwnerEmail string `json:"ownerEmail"`
}

type KuberLogicTenantStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
	Services   map[string]string  `json:"services,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=kuberlogic,scope=Cluster,shortName=klt
// +kubebuilder:subresource:status
type KuberLogicTenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KuberLogicTenantSpec   `json:"spec,omitempty"`
	Status KuberLogicTenantStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type KuberLogicTenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KuberLogicTenant `json:"items"`
}

const (
	activeCondType   = "Active"
	ktSyncedCondType = "Synced"
)

func (kt KuberLogicTenant) GetServiceAccountName() string {
	return kt.GetTenantName()
}

func (kt KuberLogicTenant) GetTenantName() string {
	return kt.ObjectMeta.Name
}

func (kt *KuberLogicTenant) SetActive() {
	kt.setConditionStatus(activeCondType, true, "Tenant is active", activeCondType)
}

func (kt KuberLogicTenant) IsActive() bool {
	return meta.IsStatusConditionTrue(kt.Status.Conditions, activeCondType)
}

func (kt *KuberLogicTenant) SetSynced() {
	kt.setConditionStatus(ktSyncedCondType, true, "Tenant is synced", ktSyncedCondType)
}

func (kt *KuberLogicTenant) SyncFailed(msg string) {
	kt.setConditionStatus(ktSyncedCondType, false, msg, ktSyncedCondType)
}

func (kt *KuberLogicTenant) IsSynced() bool {
	return meta.IsStatusConditionTrue(kt.Status.Conditions, ktSyncedCondType)
}

func (kt *KuberLogicTenant) setConditionStatus(cond string, status bool, msg, reason string) {
	c := metav1.Condition{
		Type:    cond,
		Status:  metav1.ConditionFalse,
		Message: msg,
		Reason:  reason,
	}
	if status {
		c.Status = metav1.ConditionTrue
	}
	meta.SetStatusCondition(&kt.Status.Conditions, c)
}

// SaveTenantServiceInfo saves kls name and type into a Kuberlogictenant
func (kt KuberLogicTenant) SaveTenantServiceInfo(kls *KuberLogicService) {
	if len(kt.Status.Services) == 0 {
		kt.Status.Services = make(map[string]string, 1)
	}
	kt.Status.Services[kls.Name] = kls.GetServiceType()
}

func init() {
	SchemeBuilder.Register(&KuberLogicTenant{}, &KuberLogicTenantList{})
}
