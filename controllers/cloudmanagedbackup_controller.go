package controllers

import (
	"context"
	"github.com/go-logr/logr"
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	"gitlab.com/cloudmanaged/operator/api/v1/operator"
	v12 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

// CloudManagedBackupReconciler reconciles a CloudManagedBackup object
type CloudManagedBackupReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	mu     sync.Mutex
}

// +kubebuilder:rbac:groups=cloudlinux.com,resources=cloudmanagedbackups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudlinux.com,resources=cloudmanagedbackups/status,verbs=get;update;patch
func (r *CloudManagedBackupReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("cloudmanagedbackup", req.NamespacedName)

	r.mu.Lock()
	defer r.mu.Unlock()

	// Fetch the Cloudmanaged instance
	cloudmanagedbackup := &cloudlinuxv1.CloudManagedBackup{}
	err := r.Get(ctx, req.NamespacedName, cloudmanagedbackup)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get CloudmanagedBackup")
		return ctrl.Result{}, err
	}

	clusterName := cloudmanagedbackup.Spec.ClusterName
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

	backupOperator, err := operator.GetBackupOperator(op)
	if err != nil {
		log.Error(err, "Could not define the backup operator")
		return ctrl.Result{}, err
	}
	backupOperator.SetBackupImage()
	backupOperator.SetBackupEnv(cloudmanagedbackup)

	cronJob := backupOperator.GetCronJob()
	err = r.Get(ctx,
		types.NamespacedName{
			Name:      cloudmanagedbackup.Name,
			Namespace: cloudmanagedbackup.Namespace,
		},
		cronJob)
	if err != nil && k8serrors.IsNotFound(err) {
		dep, err := r.cronJob(backupOperator, cloudmanagedbackup)
		if err != nil {
			log.Error(err, "Could not generate cron cronJob", "Name", cloudmanagedbackup.Name)
			return ctrl.Result{}, err
		}

		log.Info("Creating a new Backup resource", "Name", cloudmanagedbackup.Name)
		err = r.Create(ctx, dep)
		if err != nil && k8serrors.IsAlreadyExists(err) {
			log.Info("CronJob already exists", "Name", cloudmanagedbackup.Name)
		} else if err != nil {
			log.Error(err, "Failed to create new CronJob", "Name", cloudmanagedbackup.Name)
			return ctrl.Result{}, err
		} else {
			// cronJob created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		}
	}

	backupOperator.InitFrom(cronJob)
	if !backupOperator.IsEqual(cloudmanagedbackup) {
		backupOperator.Update(cloudmanagedbackup)

		err = r.Update(ctx, backupOperator.GetCronJob())
		if err != nil {
			log.Error(err, "Failed to update object", "Name", cloudmanagedbackup.Name)
			return ctrl.Result{}, err
		} else {
			log.Info("Backup resource is updated", "Name", cloudmanagedbackup.Name)
		}
	} else {
		log.Info("No difference", "Name", cloudmanagedbackup.Name)
	}

	jobList, err := r.getJobList(ctx, cloudmanagedbackup)
	if err != nil {
		log.Error(err, "Failed to receive list of jobs",
			"Name", cloudmanagedbackup.Name)
		return ctrl.Result{}, err
	}

	status := backupOperator.CurrentStatus(jobList)
	if !cloudmanagedbackup.IsEqual(status) {
		cloudmanagedbackup.SetStatus(status)
		err = r.Update(ctx, cloudmanagedbackup)
		//err = r.Status().Update(ctx, cloudmanaged) # FIXME: Figure out why it's failed
		if err != nil {
			log.Error(err, "Failed to update cloudmanaged backup object")
		} else {
			log.Info("Cloudmanaged backup status is updated", "Status", cloudmanagedbackup.GetStatus())
		}
	}

	return ctrl.Result{}, nil
}

func (r *CloudManagedBackupReconciler) cronJob(op operator.Backup, cmb *cloudlinuxv1.CloudManagedBackup) (*v1beta1.CronJob, error) {
	op.Init(cmb)

	// Set cloudmanage backup instance as the owner and controller
	// if cloudmanage backup will remove -> dep also should be removed automatically
	err := ctrl.SetControllerReference(cmb, op.GetCronJob(), r.Scheme)
	if err != nil {
		return nil, err
	}

	return op.GetCronJob(), nil
}

func (r *CloudManagedBackupReconciler) getJobList(ctx context.Context, cmb *cloudlinuxv1.CloudManagedBackup) (v12.JobList, error) {
	jobs := v12.JobList{}
	selector := &client.ListOptions{}

	client.InNamespace(cmb.Namespace).ApplyToList(selector)
	client.MatchingLabels{
		"backup-name": cmb.Name,
	}.ApplyToList(selector)

	err := r.List(ctx, &jobs, selector)
	return jobs, err
}

func (r *CloudManagedBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudlinuxv1.CloudManagedBackup{}).
		Owns(&v1beta1.CronJob{}).
		Complete(r)
}
