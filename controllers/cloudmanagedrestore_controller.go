package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	"gitlab.com/cloudmanaged/operator/api/v1/operator"
	"gitlab.com/cloudmanaged/operator/monitoring"
	v1 "k8s.io/api/batch/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

// CloudManagedBackupRestoreReconciler reconciles a CloudManagedRestore object
type CloudManagedRestoreReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	mu     sync.Mutex
}

// +kubebuilder:rbac:groups=cloudlinux.com,resources=cloudmanagedrestores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudlinux.com,resources=cloudmanagedrestores/status,verbs=get;update;patch
func (r *CloudManagedRestoreReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("cloudmanagedrestore", req.NamespacedName)

	r.mu.Lock()
	defer r.mu.Unlock()

	// metrics key
	monitoringKey := fmt.Sprintf("%s/%s", req.Name, req.Namespace)

	// Fetch the Cloudmanaged instance
	cloudmanagedrestore := &cloudlinuxv1.CloudManagedRestore{}
	err := r.Get(ctx, req.NamespacedName, cloudmanagedrestore)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get CloudmanagedBackup")
		delete(monitoring.CloudManageds, monitoringKey)
		return ctrl.Result{}, err
	}

	clusterName := cloudmanagedrestore.Spec.ClusterName
	cloudmanaged := &cloudlinuxv1.CloudManaged{}
	err = r.Get(
		ctx,
		types.NamespacedName{
			Name:      clusterName,
			Namespace: req.Namespace,
		},
		cloudmanaged,
	)
	// TODO: should be of part of validation CR
	if err != nil && k8serrors.IsNotFound(err) {
		log.Info("Cluster is not found",
			"Cluster", clusterName)
		return ctrl.Result{}, nil
	}

	op, err := operator.GetOperator(cloudmanaged.Spec.Type)
	if err != nil {
		log.Error(err, "Could not define the base operator")
		return ctrl.Result{}, err
	}
	found := op.AsRuntimeObject()
	err = r.Get(
		ctx,
		types.NamespacedName{
			Name:      cloudmanaged.Name,
			Namespace: cloudmanaged.Namespace,
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
	restoreOperator.SetRestoreEnv(cloudmanagedrestore)

	job := restoreOperator.GetJob()
	err = r.Get(ctx,
		types.NamespacedName{
			Name:      cloudmanagedrestore.Name,
			Namespace: cloudmanagedrestore.Namespace,
		},
		job)
	if err != nil && k8serrors.IsNotFound(err) {
		dep, err := r.defineJob(restoreOperator, cloudmanagedrestore)
		if err != nil {
			log.Error(err, "Could not generate job", "Name", cloudmanagedrestore.Name)
			return ctrl.Result{}, err
		}

		log.Info("Creating a new Restore resource", "Name", cloudmanagedrestore.Name)
		err = r.Create(ctx, dep)
		if err != nil && k8serrors.IsAlreadyExists(err) {
			log.Info("Job already exists", "Name", cloudmanagedrestore.Name)
		} else if err != nil {
			log.Error(err, "Failed to create new Job", "Name", cloudmanagedrestore.Name,
				"Namespace", cloudmanagedrestore.Namespace)
			return ctrl.Result{}, err
		} else {
			// job created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		}
	}
	restoreOperator.InitFrom(job)
	status := restoreOperator.CurrentStatus()
	if !cloudmanagedrestore.IsEqual(status) {
		cloudmanagedrestore.SetStatus(status)
		err = r.Update(ctx, cloudmanagedrestore)
		//err = r.Status().Update(ctx, cloudmanaged) # FIXME: Figure out why it's failed
		if err != nil {
			log.Error(err, "Failed to update cloudmanaged restore object")
		} else {
			log.Info("Cloudmanaged restore status is updated",
				"Status", cloudmanagedrestore.GetStatus())
		}
	}

	monitoring.CloudManagedRestores[monitoringKey] = cloudmanagedrestore

	return ctrl.Result{}, nil
}

func (r *CloudManagedRestoreReconciler) defineJob(op operator.Restore, cr *cloudlinuxv1.CloudManagedRestore) (*v1.Job, error) {

	op.Init(cr)

	// Set cloudmanage restore instance as the owner and controller
	// if cloudmanage restore will remove -> dep also should be removed automatically
	err := ctrl.SetControllerReference(cr, op.GetJob(), r.Scheme)
	if err != nil {
		return nil, err
	}

	return op.GetJob(), nil
}

func (r *CloudManagedRestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudlinuxv1.CloudManagedRestore{}).
		Owns(&v1.Job{}).
		Complete(r)
}
