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
	"github.com/pkg/errors"
	postgres "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	v12 "k8s.io/api/batch/v1"
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

func initTestKlrReconciler() *KuberLogicBackupRestoreReconciler {

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(kuberlogicv1.AddToScheme(scheme))
	utilruntime.Must(postgres.AddToScheme(scheme))
	utilruntime.Must(mysql.AddToScheme(scheme))

	b := fake.NewClientBuilder()
	c := b.WithScheme(scheme).Build()

	logger := zap.New()
	controllerruntime.SetLogger(logger)

	return NewKuberlogicBackupRestoreReconciler(
		c,
		logger.WithName("kuberlogicbackuprestore"),
		scheme,
		&cfg.Config{
			Platform:  "testing",
			Version:   "test-ver",
			ImageRepo: "quay.io",
		},
		monitoring.NewCollector(),
	)
}

func TestKuberlogicBackupRestoreReconciler_ReconcileAbsent(t *testing.T) {
	r := initTestKlrReconciler()
	klrId := types.NamespacedName{
		Name:      "absent",
		Namespace: "absent",
	}

	if _, err := r.Reconcile(context.TODO(), controllerruntime.Request{NamespacedName: klrId}); err != nil {
		t.Errorf("unexpected error %v", err)
	}
}

func TestKuberlogicBackupRestoreReconciler_ReconcileValid(t *testing.T) {
	r := initTestKlrReconciler()

	klrId := types.NamespacedName{
		Name:      "test",
		Namespace: "test",
	}
	klr := &kuberlogicv1.KuberLogicBackupRestore{
		ObjectMeta: v1.ObjectMeta{
			Name:      klrId.Name,
			Namespace: klrId.Namespace,
		},
		Spec: kuberlogicv1.KuberLogicBackupRestoreSpec{
			Type:        "mysql",
			ClusterName: klrId.Name,
			SecretName:  "fake",
			Backup:      "fake",
			Database:    "fake",
		},
	}

	if err := createKuberlogicsbackuprestoreWithEnv(r.Client, klr); err != nil {
		t.Error(err)
	}

	kls := &kuberlogicv1.KuberLogicService{
		ObjectMeta: v1.ObjectMeta{
			Name:      klrId.Name,
			Namespace: klrId.Namespace,
		},
		Spec: kuberlogicv1.KuberLogicServiceSpec{
			Type:     "mysql",
			Replicas: 1,
		},
	}
	if err := r.Client.Create(context.TODO(), kls); err != nil {
		t.Errorf("error creating test kls: %v", err)
	}
	my := &v1alpha1.MysqlCluster{
		ObjectMeta: v1.ObjectMeta{
			Name:      klrId.Name,
			Namespace: klrId.Namespace,
		},
		Spec: v1alpha1.MysqlClusterSpec{},
	}
	if err := r.Client.Create(context.TODO(), my); err != nil {
		t.Errorf("error creating test mysql: %v", err)
	}

	if _, err := r.Reconcile(context.TODO(), controllerruntime.Request{NamespacedName: klrId}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// verify that job exists
	j := &v12.Job{}
	if err := r.Client.Get(context.TODO(), klrId, j); err != nil {
		t.Errorf("error getting backup cronjob: %v", err)
	}
}

func createKuberlogicsbackuprestoreWithEnv(c client.Client, klr *kuberlogicv1.KuberLogicBackupRestore) error {
	klt := &kuberlogicv1.KuberLogicTenant{
		ObjectMeta: v1.ObjectMeta{
			Name: klr.Namespace,
		},
	}

	// typically this would be created by Klt controller
	ns := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: klr.Namespace,
		},
	}
	if err := c.Create(context.TODO(), ns); err != nil {
		return errors.Wrap(err, "error creating namespace")
	}

	if err := c.Create(context.TODO(), klt); err != nil {
		return errors.Wrap(err, "error creating tenant")
	}

	if err := c.Create(context.TODO(), klr); err != nil {
		return errors.Wrap(err, "error creating restore")
	}
	return nil
}
