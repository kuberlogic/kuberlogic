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
	"github.com/go-logr/logr"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	"github.com/kuberlogic/kuberlogic/modules/operator/cfg"
	grafana "github.com/kuberlogic/kuberlogic/modules/operator/controllers/kuberlogictenant_controller/grafana"
	"github.com/kuberlogic/kuberlogic/modules/operator/util"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sync"
)

// KuberlogicTenantReconciler reconciles a KuberlogicTenant object
type KuberlogicTenantReconciler struct {
	Client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	Config *cfg.Config
	mu     sync.Mutex
}

const (
	ktFinalizer = kuberlogicv1.Group + "/tenant-finalizer"
)

func (r *KuberlogicTenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("kuberlogictenant", req.NamespacedName) //"enabled", r.Config.Grafana.Enabled,
	//"login", r.Config.Grafana.Login,
	//"password", r.Config.Grafana.Password,
	//"endpoint", r.Config.Grafana.Endpoint,

	defer util.HandlePanic(log)

	r.mu.Lock()
	defer r.mu.Unlock()

	log.Info("reconciliation started")
	// Fetch the KuberlogicTenant instance
	kt := &kuberlogicv1.KuberLogicTenant{}
	if err := r.Client.Get(ctx, req.NamespacedName, kt); err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info(req.Namespace, req.Name, " has been deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get KuberlogicTenant")
		return ctrl.Result{}, err
	}
	if kt.DeletionTimestamp != nil {
		log.Info("kuberlogicalert is pending for deletion")
		if controllerutil.ContainsFinalizer(kt, ktFinalizer) {
			if err := finalize(ctx, r.Config, r.Client, kt, log); err != nil {
				log.Error(err, "error finalizing kuberlogicalert")
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(kt, ktFinalizer)
			if err := r.Client.Update(ctx, kt); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	}

	if !controllerutil.ContainsFinalizer(kt, ktFinalizer) {
		log.Info("adding finalizer", "finalizer", ktFinalizer)
		controllerutil.AddFinalizer(kt, ktFinalizer)
		err := r.Client.Update(ctx, kt)
		if err != nil {
			log.Error(err, "error adding finalizer")
		}
		return ctrl.Result{}, err
	}

	if r.Config.Grafana.Enabled {
		if err := grafana.NewGrafanaSyncer(kt, log, r.Config.Grafana).Sync(); err != nil {
			kt.SyncFailed(err.Error())
			log.Info("grafana sync failed", "error", err)
			return ctrl.Result{}, err
		}
	}

	err := newSyncer(ctx, log, r.Client, r.Scheme, kt).Sync(
		r.Config.ImagePullSecretName, r.Config.Namespace)
	log.Info("reconciliation finished", "error", err)

	if err == nil {
		kt.SetSynced()
		log.Info("setting object status to synced")
		return ctrl.Result{}, r.Client.Status().Update(ctx, kt)
	} else {
		kt.SyncFailed(err.Error())
		log.Info("object sync failed", "error", err)
	}

	return ctrl.Result{}, err
}

func (r *KuberlogicTenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kuberlogicv1.KuberLogicTenant{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Namespace{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Complete(r)
}
