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

package kuberlogictenant_controller

import (
	"context"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	"github.com/kuberlogic/kuberlogic/modules/operator/cfg"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"net/http"
	"net/http/httptest"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"strings"
	"sync"
	"testing"
	"time"
)

func setupTestReconciler() *KuberlogicTenantReconciler {
	config := &cfg.Config{
		Namespace: "test-ns",
	}

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(kuberlogicv1.AddToScheme(scheme))

	b := fake.NewClientBuilder()
	client := b.WithScheme(scheme).Build()

	logger := zap.New()
	controllerruntime.SetLogger(logger)

	return &KuberlogicTenantReconciler{
		Client: client,
		Log:    log.Log.WithName("TestKuberlogicTenantReconciler"),
		Scheme: scheme,
		Config: config,
		mu:     sync.Mutex{},
	}
}

func TestKuberlogicTenantReconciler_Reconcile(t *testing.T) {
	r := setupTestReconciler()

	if err := testAbsentKlt(r); err != nil {
		t.Errorf("testAbsentKlt failed with error: %v", err)
	}

	if err := testCreatedKlt(r); err != nil {
		t.Errorf("testCreatedKlt failed with error: %v", err)
	}

	if err := testKltWithGrafana(r); err != nil {
		t.Errorf("testKltWithGrafana failed with error: %v", err)
	}

	if err := testDeletionWithServices(r); err != nil {
		t.Errorf("testDeletionWithServices failed with error: %v", err)
	}
}

// testAbsentKlt tests Reconcile function when klt does not exist in cluster
func testAbsentKlt(r *KuberlogicTenantReconciler) error {
	ctx := context.TODO()
	kltId := types.NamespacedName{
		Name: "absent",
	}
	req := controllerruntime.Request{
		NamespacedName: kltId,
	}
	_, err := r.Reconcile(ctx, req)
	if err != nil {
		return errors.Wrap(err, "error reconciling object")
	}
	klt := new(kuberlogicv1.KuberLogicTenant)
	if err := r.Client.Get(ctx, kltId, klt); err != nil && !k8serrors.IsNotFound(err) {
		return errors.Wrap(err, "error checking object after reconciliation")
	}
	return nil
}

// testCreatedKlt tests a newly created kuberlogictenant
func testCreatedKlt(r *KuberlogicTenantReconciler) error {
	ctx := context.TODO()
	kltId := types.NamespacedName{
		Name: "created",
	}

	if err := r.Client.Create(ctx, &kuberlogicv1.KuberLogicTenant{
		ObjectMeta: v1.ObjectMeta{
			Name:      kltId.Name,
			Namespace: kltId.Namespace,
		},
	}); err != nil {
		return errors.Wrap(err, "error creating test klt")
	}

	req := controllerruntime.Request{
		NamespacedName: kltId,
	}
	_, err := r.Reconcile(ctx, req)
	if err != nil {
		return errors.Wrap(err, "reconciliation failed")
	}

	klt := new(kuberlogicv1.KuberLogicTenant)
	if err := r.Client.Get(ctx, kltId, klt); err != nil {
		return errors.Wrap(err, "error checking object after reconciliation")
	}
	if !klt.IsSynced() {
		return errors.New("klt must be synced")
	}

	return nil
}

// testKltWithGrafana tests
func testKltWithGrafana(r *KuberlogicTenantReconciler) error {
	grafanaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e := r.URL.Path
		// mock org create
		if strings.Contains(e, "/api/orgs/name") {
			w.WriteHeader(200)
			_, _ = w.Write([]byte("{\"id\": 1, \"name\":\"test\"}"))
		}

		// mock user search
		if strings.Contains(e, "/api/users/lookup") {
			w.WriteHeader(404)
		}

		// mock user create
		if strings.Contains(e, "/api/admin/users") {
			w.WriteHeader(200)
			_, _ = w.Write([]byte("{\"id\": 1, \"message\": \"created\"}"))
		}
	}))
	defer grafanaServer.Close()

	r.Config.Grafana = cfg.Grafana{
		Enabled:                   true,
		Endpoint:                  grafanaServer.URL + "/",
		Login:                     "user",
		Password:                  "password",
		DefaultDatasourceEndpoint: "http://localhost:1888",
	}

	ctx := context.TODO()
	kltId := types.NamespacedName{
		Name: "klt-grafana",
	}

	if err := r.Client.Create(ctx, &kuberlogicv1.KuberLogicTenant{
		ObjectMeta: v1.ObjectMeta{
			Name:      kltId.Name,
			Namespace: kltId.Namespace,
		},
		Spec: kuberlogicv1.KuberLogicTenantSpec{
			OwnerEmail: "test@example.com",
		},
	}); err != nil {
		return errors.Wrap(err, "error creating test klt")
	}

	req := controllerruntime.Request{
		NamespacedName: kltId,
	}
	_, err := r.Reconcile(ctx, req)
	if err != nil {
		return errors.Wrap(err, "reconciliation failed")
	}

	klt := new(kuberlogicv1.KuberLogicTenant)
	if err := r.Client.Get(ctx, kltId, klt); err != nil {
		return errors.Wrap(err, "error checking object after reconciliation")
	}
	if !klt.IsSynced() {
		return errors.New("klt must be synced")
	}
	r.Config.Grafana.Enabled = false
	return nil
}

// testDeletionWithServices tests klt deletion when service exist
func testDeletionWithServices(r *KuberlogicTenantReconciler) error {
	ctx := context.TODO()
	kltId := types.NamespacedName{
		Name: "klt-service-exists",
	}

	if err := r.Client.Create(ctx, &kuberlogicv1.KuberLogicTenant{
		ObjectMeta: v1.ObjectMeta{
			Name:      kltId.Name,
			Namespace: kltId.Namespace,
		},
		Status: kuberlogicv1.KuberLogicTenantStatus{
			Services: map[string]string{"my": "mysql"},
		},
	}); err != nil {
		return errors.Wrap(err, "error creating test klt")
	}

	req := controllerruntime.Request{
		NamespacedName: kltId,
	}
	_, err := r.Reconcile(ctx, req)
	if err != nil {
		return errors.Wrap(err, "reconciliation failed")
	}

	klt := new(kuberlogicv1.KuberLogicTenant)
	if err := r.Client.Get(ctx, kltId, klt); err != nil {
		return errors.Wrap(err, "error checking object after reconciliation")
	}
	if len(klt.Finalizers) == 0 {
		return errors.New("finalizers not found")
	}

	// now simulate deletion
	now := time.Now()
	klt.DeletionTimestamp = &v1.Time{Time: now}
	_ = r.Client.Update(ctx, klt)

	// and reconcile again hoping for an error
	_, err = r.Reconcile(ctx, req)
	if !errors.Is(err, errFinalizingTenant) {
		return errors.Wrap(err, "did not find expected error")
	}

	// now delete services and try again
	klt.Status.Services = make(map[string]string, 0)
	_ = r.Client.Update(ctx, klt)

	// and reconcile again hoping for an error
	_, err = r.Reconcile(ctx, req)
	if err != nil {
		return errors.Wrap(err, "object must have been deleted with errors")
	}

	klt = new(kuberlogicv1.KuberLogicTenant)
	if err := r.Client.Get(ctx, kltId, klt); err != nil {
		return errors.Wrap(err, "error checking object after reconciliation")
	}
	if len(klt.Finalizers) != 0 {
		return errors.New("object should not have any finalizers")
	}
	return nil
}
