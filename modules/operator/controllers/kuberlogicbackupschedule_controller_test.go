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
	mysql "github.com/bitpoke/mysql-operator/pkg/apis"
	"github.com/bitpoke/mysql-operator/pkg/apis/mysql/v1alpha1"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	"github.com/kuberlogic/kuberlogic/modules/operator/cfg"
	"github.com/kuberlogic/kuberlogic/modules/operator/monitoring"
	serviceoperator "github.com/kuberlogic/kuberlogic/modules/operator/service-operator"
	"github.com/pkg/errors"
	postgres "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
)

var (
	testKlbCtx = context.TODO()
)

func initTestKlbReconciler() *KuberLogicBackupScheduleReconciler {

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(kuberlogicv1.AddToScheme(scheme))
	utilruntime.Must(postgres.AddToScheme(scheme))
	utilruntime.Must(mysql.AddToScheme(scheme))

	b := fake.NewClientBuilder()
	c := b.WithScheme(scheme).Build()

	logger := zap.New()
	controllerruntime.SetLogger(logger)

	return NewKuberlogicBackupScheduleReconciler(
		c,
		logger.WithName("kuberlogicbackupschedule"),
		scheme,
		&cfg.Config{
			Platform:  "testing",
			Version:   "test-ver",
			ImageRepo: "quay.io",
		},
		monitoring.NewCollector(),
	)
}

func TestKuberlogicBackupScheduleReconciler_ReconcileAbsent(t *testing.T) {
	r := initTestKlbReconciler()
	klbId := types.NamespacedName{
		Name:      "absent",
		Namespace: "absent",
	}

	if _, err := r.Reconcile(testKlbCtx, controllerruntime.Request{NamespacedName: klbId}); err != nil {
		t.Errorf("unexpected error %v", err)
	}
}

func TestKuberlogicBackupScheduleReconciler_ReconcileAbsentKls(t *testing.T) {
	r := initTestKlbReconciler()

	klbId := types.NamespacedName{
		Name:      "absent-kls",
		Namespace: "absent-kls",
	}
	klb := &kuberlogicv1.KuberLogicBackupSchedule{
		ObjectMeta: v1.ObjectMeta{
			Name:      klbId.Name,
			Namespace: klbId.Namespace,
		},
		Spec: kuberlogicv1.KuberLogicBackupScheduleSpec{
			Type:        "mysql",
			ClusterName: "absent",
		},
	}

	if err := createKuberlogicsbackupWithEnv(r.Client, klb); err != nil {
		t.Error(err)
	}

	if _, err := r.Reconcile(testKlbCtx, controllerruntime.Request{NamespacedName: klbId}); err != errKlbServiceNotFound {
		t.Errorf("error \"%v\" does not match expected \"%v\"", err, errKlbServiceNotFound)
	}
}

func TestKuberlogicBackupScheduleReconciler_ReconcileUnknownKlsType(t *testing.T) {
	r := initTestKlbReconciler()

	klbId := types.NamespacedName{
		Name:      "absent-kls",
		Namespace: "absent-kls",
	}
	klb := &kuberlogicv1.KuberLogicBackupSchedule{
		ObjectMeta: v1.ObjectMeta{
			Name:      klbId.Name,
			Namespace: klbId.Namespace,
		},
		Spec: kuberlogicv1.KuberLogicBackupScheduleSpec{
			Type:        "mysql",
			ClusterName: klbId.Name,
		},
	}

	if err := createKuberlogicsbackupWithEnv(r.Client, klb); err != nil {
		t.Error(err)
	}

	kls := &kuberlogicv1.KuberLogicService{
		ObjectMeta: v1.ObjectMeta{
			Name:      klbId.Name,
			Namespace: klbId.Namespace,
		},
		Spec: kuberlogicv1.KuberLogicServiceSpec{
			Type:     "wrong",
			Replicas: 1,
		},
	}
	if err := r.Client.Create(testKlbCtx, kls); err != nil {
		t.Errorf("error creating test kls: %v", err)
	}

	if _, err := r.Reconcile(testKlbCtx, controllerruntime.Request{NamespacedName: klbId}); err != serviceoperator.ErrOperatorTypeUnknown {
		t.Errorf("error: %v", err)
	}
}

func TestKuberlogicBackupScheduleReconciler_ReconcileValid(t *testing.T) {
	r := initTestKlbReconciler()

	klbId := types.NamespacedName{
		Name:      "test",
		Namespace: "test",
	}
	klb := &kuberlogicv1.KuberLogicBackupSchedule{
		ObjectMeta: v1.ObjectMeta{
			Name:      klbId.Name,
			Namespace: klbId.Namespace,
		},
		Spec: kuberlogicv1.KuberLogicBackupScheduleSpec{
			Type:        "mysql",
			ClusterName: klbId.Name,
			Schedule:    "1 * * * *",
		},
	}

	if err := createKuberlogicsbackupWithEnv(r.Client, klb); err != nil {
		t.Error(err)
	}

	kls := &kuberlogicv1.KuberLogicService{
		ObjectMeta: v1.ObjectMeta{
			Name:      klbId.Name,
			Namespace: klbId.Namespace,
		},
		Spec: kuberlogicv1.KuberLogicServiceSpec{
			Type:     "mysql",
			Replicas: 1,
		},
	}
	if err := r.Client.Create(testKlbCtx, kls); err != nil {
		t.Errorf("error creating test kls: %v", err)
	}
	my := &v1alpha1.MysqlCluster{
		ObjectMeta: v1.ObjectMeta{
			Name:      klbId.Name,
			Namespace: klbId.Namespace,
		},
		Spec: v1alpha1.MysqlClusterSpec{},
	}
	if err := r.Client.Create(testKlbCtx, my); err != nil {
		t.Errorf("error creating test mysql: %v", err)
	}

	if _, err := r.Reconcile(testKlbCtx, controllerruntime.Request{NamespacedName: klbId}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// verify that cronjob exists
	cj := &batchv1beta1.CronJob{}
	if err := r.Client.Get(testKlbCtx, klbId, cj); err != nil {
		t.Errorf("error getting backup cronjob: %v", err)
	}

	if cj.Spec.Schedule != klb.Spec.Schedule {
		t.Errorf("schedules does not match")
	}
	if img := cj.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image; img != r.cfg.ImageRepo+"/backup-mysql:"+r.cfg.Version {
		t.Errorf("actual backup image (%v) does not match desired", img)
	}

	if err := r.Client.Get(testKlbCtx, klbId, klb); err != nil {
		t.Errorf("error getting kuberlogicbackupschedule")
	}

	if klb.IsRunning() {
		t.Errorf("backup is not expected to be running or finished")
	}
}

func TestKuberlogicBackupScheduleReconciler_ReconcileRunning(t *testing.T) {
	r := initTestKlbReconciler()

	klbId := types.NamespacedName{
		Name:      "test",
		Namespace: "test",
	}
	klb := &kuberlogicv1.KuberLogicBackupSchedule{
		ObjectMeta: v1.ObjectMeta{
			Name:      klbId.Name,
			Namespace: klbId.Namespace,
		},
		Spec: kuberlogicv1.KuberLogicBackupScheduleSpec{
			Type:        "mysql",
			ClusterName: klbId.Name,
			Schedule:    "1 * * * *",
		},
	}

	if err := createKuberlogicsbackupWithEnv(r.Client, klb); err != nil {
		t.Error(err)
	}

	kls := &kuberlogicv1.KuberLogicService{
		ObjectMeta: v1.ObjectMeta{
			Name:      klbId.Name,
			Namespace: klbId.Namespace,
		},
		Spec: kuberlogicv1.KuberLogicServiceSpec{
			Type:     "mysql",
			Replicas: 1,
		},
	}
	if err := r.Client.Create(testKlbCtx, kls); err != nil {
		t.Errorf("error creating test kls: %v", err)
	}
	my := &v1alpha1.MysqlCluster{
		ObjectMeta: v1.ObjectMeta{
			Name:      klbId.Name,
			Namespace: klbId.Namespace,
		},
		Spec: v1alpha1.MysqlClusterSpec{},
	}
	if err := r.Client.Create(testKlbCtx, my); err != nil {
		t.Errorf("error creating test mysql: %v", err)
	}

	j := &batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      "demo",
			Namespace: klb.Namespace,
			Labels: map[string]string{
				"backup-name": klb.Name,
			},
		},
	}
	if err := r.Client.Create(testKlbCtx, j); err != nil {
		t.Errorf("error creating backup job: %v", err)
	}

	j.Status.Active = 1
	if err := r.Client.Status().Update(testKlbCtx, j); err != nil {
		t.Errorf("error updating job status")
	}

	if _, err := r.Reconcile(testKlbCtx, controllerruntime.Request{NamespacedName: klbId}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := r.Client.Get(testKlbCtx, klbId, klb); err != nil {
		t.Errorf("error getting klb: %v", err)
	}

	if !klb.IsRunning() {
		t.Errorf("expect klb to be running")
	}
}

func TestKuberlogicBackupScheduleReconciler_ReconcileFinished(t *testing.T) {
	r := initTestKlbReconciler()

	klbId := types.NamespacedName{
		Name:      "test",
		Namespace: "test",
	}
	klb := &kuberlogicv1.KuberLogicBackupSchedule{
		ObjectMeta: v1.ObjectMeta{
			Name:      klbId.Name,
			Namespace: klbId.Namespace,
		},
		Spec: kuberlogicv1.KuberLogicBackupScheduleSpec{
			Type:        "mysql",
			ClusterName: klbId.Name,
			Schedule:    "1 * * * *",
		},
	}

	if err := createKuberlogicsbackupWithEnv(r.Client, klb); err != nil {
		t.Error(err)
	}

	kls := &kuberlogicv1.KuberLogicService{
		ObjectMeta: v1.ObjectMeta{
			Name:      klbId.Name,
			Namespace: klbId.Namespace,
		},
		Spec: kuberlogicv1.KuberLogicServiceSpec{
			Type:     "mysql",
			Replicas: 1,
		},
	}
	if err := r.Client.Create(testKlbCtx, kls); err != nil {
		t.Errorf("error creating test kls: %v", err)
	}
	my := &v1alpha1.MysqlCluster{
		ObjectMeta: v1.ObjectMeta{
			Name:      klbId.Name,
			Namespace: klbId.Namespace,
		},
		Spec: v1alpha1.MysqlClusterSpec{},
	}
	if err := r.Client.Create(testKlbCtx, my); err != nil {
		t.Errorf("error creating test mysql: %v", err)
	}

	j := &batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      "demo",
			Namespace: klb.Namespace,
			Labels: map[string]string{
				"backup-name": klb.Name,
			},
		},
		Status: batchv1.JobStatus{
			Active: 1,
		},
	}
	if err := r.Client.Create(testKlbCtx, j); err != nil {
		t.Errorf("error creating backup job: %v", err)
	}

	if _, err := r.Reconcile(testKlbCtx, controllerruntime.Request{NamespacedName: klbId}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := r.Client.Get(testKlbCtx, klbId, klb); err != nil {
		t.Errorf("error getting klb: %v", err)
	}

	if !klb.IsRunning() {
		t.Errorf("backup must be running")
	}

	j.Status.Active = 0
	j.Status.Conditions = append(j.Status.Conditions, batchv1.JobCondition{Type: batchv1.JobComplete})
	if err := r.Client.Status().Update(testKlbCtx, j); err != nil {
		t.Errorf("error updating job status")
	}

	if _, err := r.Reconcile(testKlbCtx, controllerruntime.Request{NamespacedName: klbId}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := r.Client.Get(testKlbCtx, klbId, klb); err != nil {
		t.Errorf("error getting klb: %v", err)
	}

	if klb.IsRunning() {
		t.Errorf("do not expect klb to be running")
	}
	if !klb.IsFailed() {
		t.Errorf("expect klb to be succesful")
	}
}

func createKuberlogicsbackupWithEnv(c client.Client, klb *kuberlogicv1.KuberLogicBackupSchedule) error {
	klt := &kuberlogicv1.KuberLogicTenant{
		ObjectMeta: v1.ObjectMeta{
			Name: klb.Namespace,
		},
	}

	// typically this would be created by Klt controller
	ns := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: klb.Namespace,
		},
	}
	if err := c.Create(context.TODO(), ns); err != nil {
		return errors.Wrap(err, "error creating namespace")
	}

	if err := c.Create(context.TODO(), klt); err != nil {
		return errors.Wrap(err, "error creating tenant")
	}

	if err := c.Create(context.TODO(), klb); err != nil {
		return errors.Wrap(err, "error creating backupschedule")
	}
	return nil
}
