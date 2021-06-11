package kuberlogictenant_controller

import (
	"context"
	"github.com/go-logr/logr"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/cfg"
	"github.com/kuberlogic/operator/modules/operator/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/rbac/v1beta1"
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

// reconciliationResult indicates what changes have been made during reconciliation
type reconciliationResult struct {
	exit bool
	err error
}

const (
	ktFinalizer = kuberlogicv1.Group + "/tenant-finalizer"
)

func (r *KuberlogicTenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("kuberlogictenant", req.NamespacedName)
	client := r.Client

	defer util.HandlePanic(log)

	r.mu.Lock()
	defer r.mu.Unlock()

	log.Info("reconciliation started")
	// Fetch the KuberlogicTenant instance
	kt := &kuberlogicv1.KuberLogicTenant{}
	if err := client.Get(ctx, req.NamespacedName, kt); err != nil {
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
			if err := finalize(ctx, client, kt, log); err != nil {
				log.Error(err, "error finalizing kuberlogicalert")
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(kt, ktFinalizer)
			if err := client.Update(ctx, kt); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	}

	if !controllerutil.ContainsFinalizer(kt, ktFinalizer) {
		log.Info("adding finalizer", "finalizer", ktFinalizer)
		controllerutil.AddFinalizer(kt, ktFinalizer)
		err := client.Update(ctx, kt)
		if err != nil {
			log.Error(err, "error adding finalizer")
		}
		return ctrl.Result{}, err
	}

	var syncErr error
	s := newSyncer(ctx, log, r.Client, r.Scheme, kt, syncErr).
		withNamespace().
		withImagePullSecret(r.Config.ImagePullSecretName, r.Config.Namespace).
		withServiceAccount()
	log.Info("reconciliation finished", "error", s.syncErr)

	return ctrl.Result{}, s.syncErr
}

func (r *KuberlogicTenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kuberlogicv1.KuberLogicTenant{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Namespace{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&v1beta1.Role{}).
		Owns(&v1beta1.RoleBinding{}).
		Complete(r)
}
