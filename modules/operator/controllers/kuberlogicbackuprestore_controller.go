package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/monitoring"
	serviceOperator "github.com/kuberlogic/operator/modules/operator/service-operator"
	"github.com/kuberlogic/operator/modules/operator/service-operator/interfaces"
	"github.com/kuberlogic/operator/modules/operator/util"
	v1 "k8s.io/api/batch/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sync"
)

// KuberLogicBackupRestoreReconciler reconciles a KuberLogicBackupRestore object
type KuberLogicBackupRestoreReconciler struct {
	client.Client
	Log                 logr.Logger
	Scheme              *runtime.Scheme
	mu                  sync.Mutex
	MonitoringCollector *monitoring.KuberLogicCollector
}

const (
	backupRestoreFinalizer = "/backuprestore-finalizer"
)

// +kubebuilder:rbac:groups=cloudlinux.com,resources=kuberlogicbackuprestores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudlinux.com,resources=kuberlogicbackuprestores/status,verbs=get;update;patch
func (r *KuberLogicBackupRestoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("kuberlogicbackuprestore", req.NamespacedName)

	defer util.HandlePanic(log)

	mu := getMutex(req.NamespacedName)
	mu.Lock()
	defer mu.Unlock()

	// metrics key
	monitoringKey := fmt.Sprintf(req.NamespacedName.String())

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
		r.MonitoringCollector.ForgetKuberlogicRestore(monitoringKey)
		return ctrl.Result{}, err
	}
	defer r.MonitoringCollector.MonitorKuberlogicRestore(monitoringKey, klr)

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

	if klr.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(klr, backupRestoreFinalizer) {
			if err := r.finalize(ctx, kl, log); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(klr, backupRestoreFinalizer)
			if err := r.Update(ctx, klr); err != nil {
				log.Error(err, "error removing finalizer")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	}

	// add finalizer if it is not already there
	if controllerutil.ContainsFinalizer(klr, backupRestoreFinalizer) {
		controllerutil.AddFinalizer(klr, backupRestoreFinalizer)
		if err := r.Update(ctx, klr); err != nil {
			log.Error(err, "error adding finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// get tenant information first
	kt := &kuberlogicv1.KuberLogicTenant{}
	if err := r.Get(ctx, types.NamespacedName{Name: klr.Namespace}, kt); err != nil {
		log.Error(err, "error searching for kuberlogic tenant", "name", klr.Namespace)
		return ctrl.Result{}, err
	}

	op, err := serviceOperator.GetOperator(kl.Spec.Type)
	if err != nil {
		log.Error(err, "Could not define the base operator")
		return ctrl.Result{}, err
	}
	found := op.AsClientObject()
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

	backupRestore := op.GetBackupRestore()
	backupRestore.SetRestoreImage()
	backupRestore.SetRestoreEnv(klr)
	backupRestore.SetServiceAccount(kt.GetServiceAccountName())

	job := backupRestore.GetJob()
	err = r.Get(ctx,
		types.NamespacedName{
			Name:      klr.Name,
			Namespace: klr.Namespace,
		},
		job)
	if err != nil && k8serrors.IsNotFound(err) {
		dep, err := r.defineJob(ctx, backupRestore, klr, log)
		if err != nil {
			log.Error(err, "Could not generate job", "Name", klr.Name)
			return ctrl.Result{}, err
		}

		log.Info("Creating a new BackupRestore resource", "Name", klr.Name)
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
	backupRestore.InitFrom(job)

	if running := backupRestore.IsRunning(); running {
		klr.MarkRunning()

		// also notify corresponding kls that it is running
		kl.RestoreStarted(job.Name)
		if err := r.Status().Update(ctx, kl); err != nil {
			log.Error(err, "error updating kuberlogicservice restore condition")
			return ctrl.Result{}, err
		}
	} else {
		kl.RestoreFinished()
		if err := r.Status().Update(ctx, kl); err != nil {
			log.Error(err, "error updating kuberlogicservice restore condition")
			return ctrl.Result{}, err
		}
	}
	if successful := backupRestore.IsSuccessful(); successful {
		klr.MarkSuccessfulFinish()
	} else {
		klr.MarkFailed()
	}

	err = r.Status().Update(ctx, klr)
	if err != nil {
		log.Error(err, "Failed to update kl restore status")
		return ctrl.Result{}, err
	}
	log.Info("KuberLogicBackupRestore status is updated")

	return ctrl.Result{}, nil
}

func (r *KuberLogicBackupRestoreReconciler) defineJob(ctx context.Context, op interfaces.BackupRestore, cr *kuberlogicv1.KuberLogicBackupRestore, log logr.Logger) (*v1.Job, error) {
	// Set kuberlogic restore instance as the owner and controller
	// if kuberlogic restore will remove -> dep also should be removed automatically
	err := ctrl.SetControllerReference(cr, op.GetJob(), r.Scheme)
	if err != nil {
		return nil, err
	}

	return op.GetJob(), nil
}

func (r *KuberLogicBackupRestoreReconciler) finalize(ctx context.Context, kl *kuberlogicv1.KuberLogicService, log logr.Logger) error {
	log.Info("finalizing backuprestore")
	kl.RestoreFinished()
	return r.Status().Update(ctx, kl)
}

func (r *KuberLogicBackupRestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kuberlogicv1.KuberLogicBackupRestore{}).
		Owns(&v1.Job{}).
		Complete(r)
}
