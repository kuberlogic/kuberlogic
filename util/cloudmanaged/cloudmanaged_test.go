package cloudmanaged

import (
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
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

func TestGetClusterPodLabels(t *testing.T) {
	masterPgE, replicaPgE :=
		map[string]string{"spilo-role": "master", "application": "spilo", "cluster-name": "test-pg"},
		map[string]string{"spilo-role": "replica", "application": "spilo", "cluster-name": "test-pg"}
	masterPgA, replicaPgA, _ := GetClusterPodLabels(cmPg)

	if !reflect.DeepEqual(masterPgA, masterPgE) || !reflect.DeepEqual(replicaPgA, replicaPgE) {
		t.Errorf("replica or password pod labels are not correct for %s of type %s.\nActual: %v, %v\nExpected: %v, %v",
			cmPg.Name, cmPg.Spec.Type,
			masterPgA, replicaPgA,
			masterPgE, replicaPgE)
	}

	masterMyE, replicaMyE :=
		map[string]string{"role": "master", "mysql.presslabs.org/cluster": "test-mysql"},
		map[string]string{"role": "replica", "mysql.presslabs.org/cluster": "test-mysql"}
	masterMyA, replicaMyA, _ := GetClusterPodLabels(cmMy)

	if !reflect.DeepEqual(masterMyA, masterMyE) || !reflect.DeepEqual(replicaMyA, replicaMyE) {
		t.Errorf("replica or password pod labels are not correct for %s of type %s.\nActual: %v, %v\nExpected: %v, %v",
			cmMy.Name, cmMy.Spec.Type,
			masterMyA, replicaMyA,
			masterMyE, replicaMyE)
	}
}
