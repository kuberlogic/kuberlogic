package cloudmanaged

import (
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

var cmPg = &cloudlinuxv1.CloudManaged{
	TypeMeta: metav1.TypeMeta{},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-pg",
		Namespace: "default",
	},
	Spec: cloudlinuxv1.CloudManagedSpec{
		Type:         "postgresql",
		Replicas:     2,
		SecretName:   "",
		Resources:    v1.ResourceRequirements{},
		VolumeSize:   "1Gi",
		Version:      "11.1",
		AdvancedConf: map[string]string{"k1": "v1"},
		MaintenanceWindow: cloudlinuxv1.MaintenanceWindow{
			StartHour: 4,
			Weekday:   "Monday",
		},
		DefaultUser: "cloudmanaged",
	},
	Status: cloudlinuxv1.CloudManagedStatus{
		Status: "Running",
	},
}

var cmMy = &cloudlinuxv1.CloudManaged{
	TypeMeta: metav1.TypeMeta{},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-mysql",
		Namespace: "default",
	},
	Spec: cloudlinuxv1.CloudManagedSpec{
		Type:         "mysql",
		Replicas:     2,
		SecretName:   "dumb-secret",
		Resources:    v1.ResourceRequirements{},
		VolumeSize:   "1Gi",
		Version:      "5.11",
		AdvancedConf: map[string]string{"k2": "v2", "k3": "5"},
		MaintenanceWindow: cloudlinuxv1.MaintenanceWindow{
			StartHour: 5,
			Weekday:   "Sunday",
		},
		DefaultUser: "cloudmanaged",
	},
	Status: cloudlinuxv1.CloudManagedStatus{
		Status: "Running",
	},
}

func TestGetClusterCredentialsInfo(t *testing.T) {
	pgUserE, pgPasswordFieldE, pgSecretE := "cloudmanaged", "password", "cloudmanaged.test-pg.credentials"
	pgUserA, pgPasswordFieldA, pgSecretA, _ := GetClusterCredentialsInfo(cmPg)

	if pgUserE != pgUserA || pgPasswordFieldE != pgPasswordFieldA || pgSecretE != pgSecretA {
		t.Errorf("user or passwordfield or secret is not correct for %s of type %s.\nActual: %s, %s, %s\nExpected: %s, %s, %s",
			cmPg.Name, cmPg.Spec.Type,
			pgUserA, pgPasswordFieldA, pgSecretA,
			pgUserE, pgPasswordFieldE, pgSecretE)
	}

	myUserE, myPasswordFieldE, mySecretE := "cloudmanaged", "PASSWORD", "dumb-secret"
	myUserA, myPasswordFieldA, mySecretA, _ := GetClusterCredentialsInfo(cmMy)

	if myUserA != myUserE || myPasswordFieldA != myPasswordFieldE || mySecretA != mySecretE {
		t.Errorf("user or passwordfield or secret is not correct for %s of type %s.\nActual: %s, %s, %s\nExpected: %s, %s, %s",
			cmMy.Name, cmMy.Spec.Type,
			myUserA, myPasswordFieldA, mySecretA,
			myUserE, myPasswordFieldE, mySecretE)
	}
}
