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

package controllers

import (
	"context"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	"github.com/kuberlogic/kuberlogic/modules/operator/monitoring"
	serviceoperator "github.com/kuberlogic/kuberlogic/modules/operator/service-operator"
	"github.com/pkg/errors"
	mysql "github.com/presslabs/mysql-operator/pkg/apis"
	"github.com/presslabs/mysql-operator/pkg/apis/mysql/v1alpha1"
	postgres "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sync"
	"testing"
)

var (
	testKlsSkeleton = &kuberlogicv1.KuberLogicService{
		Spec: kuberlogicv1.KuberLogicServiceSpec{
			Type:       "postgresql",
			Replicas:   3,
			VolumeSize: "1G",
			Resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("0.1"),
					corev1.ResourceMemory: resource.MustParse("1G"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("0.5"),
					corev1.ResourceMemory: resource.MustParse("2G"),
				},
			},
		},
	}
)

func initTestReconciler() *KuberLogicServiceReconciler {

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(kuberlogicv1.AddToScheme(scheme))
	utilruntime.Must(postgres.AddToScheme(scheme))
	utilruntime.Must(mysql.AddToScheme(scheme))

	b := fake.NewClientBuilder()
	c := b.WithScheme(scheme).Build()

	logger := zap.New()
	controllerruntime.SetLogger(logger)

	return &KuberLogicServiceReconciler{
		Client:              c,
		Log:                 logger.WithName("kuberlogicservice"),
		Scheme:              scheme,
		mu:                  sync.Mutex{},
		MonitoringCollector: monitoring.NewCollector(),
	}
}

func TestKlsReconcileAbsent(t *testing.T) {
	r := initTestReconciler()

	ctx := context.TODO()
	klsId := types.NamespacedName{
		Name:      "absent",
		Namespace: "absent",
	}

	req := controllerruntime.Request{
		NamespacedName: klsId,
	}
	if _, err := r.Reconcile(ctx, req); err != nil {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestKlsReconcileUnknownType(t *testing.T) {
	r := initTestReconciler()

	ctx := context.TODO()
	klsId := types.NamespacedName{
		Name:      "unknown-type",
		Namespace: "unknown-type",
	}

	kls := &kuberlogicv1.KuberLogicService{
		ObjectMeta: v1.ObjectMeta{
			Name:      klsId.Name,
			Namespace: klsId.Namespace,
		},
		Spec: kuberlogicv1.KuberLogicServiceSpec{Type: "unknown"},
	}
	if err := createKuberlogicserviceWithEnv(r.Client, kls); err != nil {
		t.Fatalf("error creating kls: %v", err)
	}

	req := controllerruntime.Request{
		NamespacedName: klsId,
	}
	_, err := r.Reconcile(ctx, req)
	if err != serviceoperator.ErrOperatorTypeUnknown {
		t.Fatalf("error %v does not match expected %v", err, serviceoperator.ErrOperatorTypeUnknown)
	}
}

func TestKlsReconcileMysql(t *testing.T) {
	r := initTestReconciler()

	ctx := context.TODO()
	klsId := types.NamespacedName{
		Name:      "valid-mysql",
		Namespace: "valid-mysql",
	}

	kls := testKlsSkeleton
	kls.ObjectMeta = v1.ObjectMeta{
		Name:      klsId.Name,
		Namespace: klsId.Namespace,
	}
	kls.Spec.Type = "mysql"

	if err := createKuberlogicserviceWithEnv(r.Client, kls); err != nil {
		t.Fatalf("error creating kuberlogicservice: %v", err)
	}

	req := controllerruntime.Request{
		NamespacedName: klsId,
	}
	_, err := r.Reconcile(ctx, req)
	if err != nil {
		t.Fatalf("error reconciling kuberlogicservice: %v", err)
	}

	mysqlOp := &v1alpha1.MysqlCluster{}
	if err := r.Get(ctx, klsId, mysqlOp); err != nil {
		t.Fatalf("error getting mysql resource %v", err)
	}
	if kls.Spec.Replicas != *mysqlOp.Spec.Replicas {
		t.Fatalf("mysql / kls replicas numbers are not equal")
	}
}

func TestKlsReconcilePsql(t *testing.T) {
	r := initTestReconciler()

	ctx := context.TODO()
	klsId := types.NamespacedName{
		Name:      "valid-psql",
		Namespace: "valid-psql",
	}
	kls := testKlsSkeleton
	kls.ObjectMeta = v1.ObjectMeta{
		Name:      klsId.Name,
		Namespace: klsId.Namespace,
	}
	kls.Spec.Type = "postgresql"

	if err := createKuberlogicserviceWithEnv(r.Client, kls); err != nil {
		t.Fatalf("error creating kuberlogicservice: %v", err)
	}

	req := controllerruntime.Request{
		NamespacedName: klsId,
	}
	_, err := r.Reconcile(ctx, req)
	if err != nil {
		t.Fatalf("error reconciling kuberlogicservice: %v", err)
	}

	op := &postgres.Postgresql{}
	// it is different from kls for postgres and includes kuberlogic as a team id
	opId := types.NamespacedName{
		Name:      "kuberlogic-" + klsId.Name,
		Namespace: klsId.Namespace,
	}
	if err := r.Get(ctx, opId, op); err != nil {
		t.Fatalf("error getting kuberlogicservice: %v", err)
	}
	if kls.Spec.Replicas != op.Spec.NumberOfInstances {
		t.Fatalf("postgres / kls replicas number are not equal")
	}
}

func TestKlsReconcileNotReady(t *testing.T) {
	r := initTestReconciler()

	ctx := context.TODO()
	klsId := types.NamespacedName{
		Name:      "valid-postgres",
		Namespace: "valid-postgres",
	}
	kls := testKlsSkeleton
	kls.ObjectMeta = v1.ObjectMeta{
		Name:      klsId.Name,
		Namespace: klsId.Namespace,
	}

	if err := createKuberlogicserviceWithEnv(r.Client, kls); err != nil {
		t.Fatalf("error creating kuberlogicservice: %v", err)
	}

	req := controllerruntime.Request{
		NamespacedName: klsId,
	}

	if err := r.Client.Get(ctx, klsId, kls); err != nil {
		t.Fatalf("error getting kuberlogicservice: %v", err)
	}

	if _, err := r.Reconcile(ctx, req); err != nil {
		t.Fatalf("did not expect an error during the first run: %v", err)
	}

	if _, err := r.Reconcile(ctx, req); err != errKlsNotReady {
		t.Fatalf("error \"%v\" does not match expected \"%v\"", err, errKlsNotReady)
	}
}

func TestKlsReconcileUpdate(t *testing.T) {
	r := initTestReconciler()

	ctx := context.TODO()
	// we'll change replicas for mysql resource
	klsId := types.NamespacedName{
		Name:      "valid-mysql",
		Namespace: "valid-mysql",
	}
	kls := testKlsSkeleton
	kls.ObjectMeta = v1.ObjectMeta{
		Name:      klsId.Name,
		Namespace: klsId.Namespace,
	}
	kls.Spec.Type = "mysql"

	if err := createKuberlogicserviceWithEnv(r.Client, kls); err != nil {
		t.Fatalf("error creating kuberlogicservice: %v", err)
	}

	if _, err := r.Reconcile(ctx, controllerruntime.Request{NamespacedName: klsId}); err != nil {
		t.Fatalf("error reconciling kuberlogicservice: %v", err)
	}

	op := new(v1alpha1.MysqlCluster)
	if err := r.Client.Get(ctx, klsId, op); err != nil {
		t.Fatalf("error getting mysql resource: %v", err)
	}
	op.Status.Conditions = append(op.Status.Conditions, v1alpha1.ClusterCondition{
		Type:    v1alpha1.ClusterConditionReady,
		Status:  "True",
		Message: "tests",
		Reason:  "tests",
	})
	if err := r.Client.Status().Update(ctx, op); err != nil {
		t.Fatalf("error settings operator status to true; %v", err)
	}

	if err := r.Client.Get(ctx, klsId, kls); err != nil {
		t.Fatalf("error getting kuberlogicservice: %v", err)
	}

	kls.Spec.Replicas += 1
	if err := r.Client.Update(ctx, kls); err != nil {
		t.Fatalf("error updating kuberlogicservice: %v", err)
	}

	if _, err := r.Reconcile(ctx, controllerruntime.Request{NamespacedName: klsId}); err != nil {
		t.Fatalf("error reconciling kuberlogicservice: %v", err)
	}

	op = new(v1alpha1.MysqlCluster)
	if err := r.Client.Get(ctx, klsId, op); err != nil {
		t.Fatalf("error getting mysql resource: %v", err)
	}
	if kls.Spec.Replicas != *op.Spec.Replicas {
		t.Fatalf("mysql / kls replicas number are not equal")
	}
}

func createKuberlogicserviceWithEnv(c client.Client, kls *kuberlogicv1.KuberLogicService) error {
	klt := &kuberlogicv1.KuberLogicTenant{
		ObjectMeta: v1.ObjectMeta{
			Name: kls.Namespace,
		},
	}

	// typically this would be created by Klt controller
	ns := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: kls.Namespace,
		},
	}
	if err := c.Create(context.TODO(), ns); err != nil {
		return errors.Wrap(err, "error creating namespace")
	}

	if err := c.Create(context.TODO(), klt); err != nil {
		return errors.Wrap(err, "error creating tenant")
	}

	if err := c.Create(context.TODO(), kls); err != nil {
		return errors.Wrap(err, "error creating service")
	}
	return nil
}
