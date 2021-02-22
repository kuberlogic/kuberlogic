package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	kuberlogicv1 "github.com/kuberlogic/operator/api/v1"
	"github.com/kuberlogic/operator/api/v1/operator"
	"github.com/kuberlogic/operator/monitoring"
	v1 "k8s.io/api/batch/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

// KuberLogicBackupRestoreReconciler reconciles a KuberLogicBackupRestore object
type KuberLogicBackupRestoreReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	mu     sync.Mutex
}

// +kubebuilder:rbac:groups=cloudlinux.com,resources=kuberlogicbackuprestores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudlinux.com,resources=kuberlogicbackuprestores/status,verbs=get;update;patch
func (r *KuberLogicBackupRestoreReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("kuberlogicbackuprestore", req.NamespacedName)

	r.mu.Lock()
	defer r.mu.Unlock()

	// metrics key
	monitoringKey := fmt.Sprintf("%s/%s", req.Name, req.Namespace)

	// Fetch the KuberLogicBackupRestore instance
	klr := &kuberlogicv1.KuberLogicBackupRestore{}
	err := r.Get(ctx, req.NamespacedName, klr)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get KuberLogicBackupRestore")
		delete(monitoring.KuberLogicServices, monitoringKey)
		return ctrl.Result{}, err
	}

	clusterName := klr.Spec.ClusterName
	kl := &kuberlogicv1.KuberLogicService{}
	err = r.Get(
		ctx,
		types.NamespacedName{
			Name:      clusterName,
			Namespace: req.Namespace,
		},
		kl,
	)
	// TODO: should be of part of validation CR
	if err != nil && k8serrors.IsNotFound(err) {
		log.Info("Cluster is not found",
			"Cluster", clusterName)
		return ctrl.Result{}, nil
	}

	op, err := operator.GetOperator(kl.Spec.Type)
	if err != nil {
		log.Error(err, "Could not define the base operator")
		return ctrl.Result{}, err
	}
	found := op.AsRuntimeObject()
	err = r.Get(
		ctx,
		types.NamespacedName{
			Name:      op.Name(kl),
			Namespace: kl.Namespace,
		},
		found,
	)
	if err != nil {
		log.Error(err, "Base operator not found")
		return ctrl.Result{}, err
	}
	op.InitFrom(found)

	restoreOperator, err := operator.GetRestoreOperator(op)
	if err != nil {
		log.Error(err, "Could not define the backup operator")
		return ctrl.Result{}, err
	}
	restoreOperator.SetRestoreImage()
	restoreOperator.SetRestoreEnv(klr)

	job := restoreOperator.GetJob()
	err = r.Get(ctx,
		types.NamespacedName{
			Name:      klr.Name,
			Namespace: klr.Namespace,
		},
		job)
	if err != nil && k8serrors.IsNotFound(err) {
		dep, err := r.defineJob(restoreOperator, klr)
		if err != nil {
			log.Error(err, "Could not generate job", "Name", klr.Name)
			return ctrl.Result{}, err
		}

		log.Info("Creating a new Restore resource", "Name", klr.Name)
		err = r.Create(ctx, dep)
		if err != nil && k8serrors.IsAlreadyExists(err) {
			log.Info("Job already exists", "Name", klr.Name)
		} else if err != nil {
			log.Error(err, "Failed to create new Job", "Name", klr.Name,
				"Namespace", klr.Namespace)
			return ctrl.Result{}, err
		} else {
			// job created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		}
	}
	restoreOperator.InitFrom(job)
	status := restoreOperator.CurrentStatus()
	if !klr.IsEqual(status) {
		klr.SetStatus(status)
		err = r.Update(ctx, klr)
		//err = r.Status().Update(ctx, kl) # FIXME: Figure out why it's failed
		if err != nil {
			log.Error(err, "Failed to update kl restore object")
		} else {
			log.Info("KuberLogicBackupRestore status is updated",
				"Status", klr.GetStatus())
		}
	}

	monitoring.KuberLogicBackupRestores[monitoringKey] = klr

	return ctrl.Result{}, nil
}

func (r *KuberLogicBackupRestoreReconciler) defineJob(op operator.Restore, cr *kuberlogicv1.KuberLogicBackupRestore) (*v1.Job, error) {

	op.Init(cr)

	// Set cloudmanage restore instance as the owner and controller
	// if cloudmanage restore will remove -> dep also should be removed automatically
	err := ctrl.SetControllerReference(cr, op.GetJob(), r.Scheme)
	if err != nil {
		return nil, err
	}

	return op.GetJob(), nil
}

func (r *KuberLogicBackupRestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kuberlogicv1.KuberLogicBackupRestore{}).
		Owns(&v1.Job{}).
		Complete(r)
}
