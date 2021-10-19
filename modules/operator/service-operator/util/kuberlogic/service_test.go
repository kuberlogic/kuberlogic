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

package kuberlogic

import (
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
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

func TestGetClusterServices(t *testing.T) {
	pgMasterSvcE, pgReplicaSvcE := "kuberlogic-test-pg", "kuberlogic-test-pg-repl"
	pgMasterSvcA, pgReplicaSvcA, _ := GetClusterServices(cmPg)

	if pgMasterSvcE != pgMasterSvcA || pgReplicaSvcE != pgReplicaSvcA {
		t.Errorf("master and/or replica service are not correct for %s of type %s.\nActual: %s, %s\nExpected: %s, %s",
			cmPg.Name, cmPg.Spec.Type,
			pgMasterSvcA, pgReplicaSvcA,
			pgReplicaSvcE, pgReplicaSvcE)
	}

	mysqlMasterSvcE, mysqlReplicaSvcE := "test-mysql-mysql-master", "test-mysql-mysql-replicas"
	mysqlMasterSvcA, mysqlReplicaSvcA, _ := GetClusterServices(cmMy)

	if mysqlMasterSvcE != mysqlMasterSvcA || mysqlReplicaSvcE != mysqlReplicaSvcA {
		t.Errorf("master and/or replica service are not correct for %s of type %s.\nActual: %s, %s\nExpected: %s, %s",
			cmMy.Name, cmMy.Spec.Type,
			mysqlMasterSvcA, mysqlReplicaSvcA,
			mysqlReplicaSvcE, mysqlReplicaSvcE)
	}
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

	myUserE, myPasswordFieldE, mySecretE := "root", "ROOT_PASSWORD", "test-mysql-cred"
	myUserA, myPasswordFieldA, mySecretA, _ := GetClusterCredentialsInfo(cmMy)

	if myUserA != myUserE || myPasswordFieldA != myPasswordFieldE || mySecretA != mySecretE {
		t.Errorf("user or passwordfield or secret is not correct for %s of type %s.\nActual: %s, %s, %s\nExpected: %s, %s, %s",
			cmMy.Name, cmMy.Spec.Type,
			myUserA, myPasswordFieldA, mySecretA,
			myUserE, myPasswordFieldE, mySecretE)
	}
}
