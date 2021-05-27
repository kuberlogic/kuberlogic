package kuberlogic

import (
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

var cmPg = &kuberlogicv1.KuberLogicService{
	TypeMeta: metav1.TypeMeta{},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-pg",
		Namespace: "default",
	},
	Spec: kuberlogicv1.KuberLogicServiceSpec{
		Type:         "postgresql",
		Replicas:     2,
		Resources:    v1.ResourceRequirements{},
		VolumeSize:   "1Gi",
		Version:      "11.1",
		AdvancedConf: map[string]string{"k1": "v1"},
		MaintenanceWindow: kuberlogicv1.MaintenanceWindow{
			StartHour: 4,
			Weekday:   "Monday",
		},
	},
	Status: kuberlogicv1.KuberLogicServiceStatus{},
}

var cmMy = &kuberlogicv1.KuberLogicService{
	TypeMeta: metav1.TypeMeta{},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-mysql",
		Namespace: "default",
	},
	Spec: kuberlogicv1.KuberLogicServiceSpec{
		Type:         "mysql",
		Replicas:     2,
		Resources:    v1.ResourceRequirements{},
		VolumeSize:   "1Gi",
		Version:      "5.11",
		AdvancedConf: map[string]string{"k2": "v2", "k3": "5"},
		MaintenanceWindow: kuberlogicv1.MaintenanceWindow{
			StartHour: 5,
			Weekday:   "Sunday",
		},
	},
	Status: kuberlogicv1.KuberLogicServiceStatus{},
}

func TestGetClusterCredentialsInfo(t *testing.T) {
	pgUserE, pgPasswordFieldE, pgSecretE := "kuberlogic", "password", "kuberlogic.kuberlogic-test-pg.credentials"
	pgUserA, pgPasswordFieldA, pgSecretA, _ := GetClusterCredentialsInfo(cmPg)

	if pgUserE != pgUserA || pgPasswordFieldE != pgPasswordFieldA || pgSecretE != pgSecretA {
		t.Errorf("user or passwordfield or secret is not correct for %s of type %s.\nActual: %s, %s, %s\nExpected: %s, %s, %s",
			cmPg.Name, cmPg.Spec.Type,
			pgUserA, pgPasswordFieldA, pgSecretA,
			pgUserE, pgPasswordFieldE, pgSecretE)
	}

	myUserE, myPasswordFieldE, mySecretE := "kuberlogic", "PASSWORD", "test-mysql-cred"
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
		map[string]string{"spilo-role": "master", "application": "spilo", "cluster-name": "kuberlogic-test-pg"},
		map[string]string{"spilo-role": "replica", "application": "spilo", "cluster-name": "kuberlogic-test-pg"}
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
