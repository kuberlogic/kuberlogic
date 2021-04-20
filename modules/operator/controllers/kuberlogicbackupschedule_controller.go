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
	v12 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sort"
	"sync"
)

// KuberLogicBackupScheduleReconciler reconciles a KuberLogicBackupSchedule object
type KuberLogicBackupScheduleReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	mu     sync.Mutex
}

const (
	backupScheduleFinalizer = "kuberlogic.com/backupschedule-finalizer"
)

// +kubebuilder:rbac:groups=cloudlinux.com,resources=kuberlogicbackupschedules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudlinux.com,resources=kuberlogicbackupschedules/status,verbs=get;update;patch
func (r *KuberLogicBackupScheduleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("kuberlogicbackupschedule", req.NamespacedName)

	defer util.HandlePanic(log)

	mu := getMutex(req.NamespacedName)
	mu.Lock()
	defer mu.Unlock()

	// metrics key
	monitoringKey := fmt.Sprintf("%s/%s", req.Name, req.Namespace)

	// Fetch the KuberLogicBackupSchedule instance
	klb := &kuberlogicv1.KuberLogicBackupSchedule{}
	err := r.Get(ctx, req.NamespacedName, klb)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get KuberLogicBackupSchedule")
		delete(monitoring.KuberLogicServices, monitoringKey)
		return ctrl.Result{}, err
	}
	monitoring.KuberLogicBackupSchedules[monitoringKey] = klb

	clusterName := klb.Spec.ClusterName
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

	// check if klb is about to be deleted and we should finalize it
	if klb.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(klb, backupScheduleFinalizer) {
			if err := r.finalize(ctx, kl, log); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(klb, backupScheduleFinalizer)
			if err := r.Update(ctx, klb); err != nil {
				log.Error(err, "error removing finalizer")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	}

	// add finalizer if it is not already there
	if controllerutil.ContainsFinalizer(klb, backupScheduleFinalizer) {
		controllerutil.AddFinalizer(klb, backupScheduleFinalizer)
		if err := r.Update(ctx, klb); err != nil {
			log.Error(err, "error adding finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
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

	backupSchedule := op.GetBackupSchedule()
	backupSchedule.SetBackupImage()
	backupSchedule.SetBackupEnv(klb)

	cronJob := backupSchedule.GetCronJob()
	err = r.Get(ctx,
		types.NamespacedName{
			Name:      klb.Name,
			Namespace: klb.Namespace,
		},
		cronJob)
	if err != nil && k8serrors.IsNotFound(err) {
		dep, err := r.cronJob(backupSchedule, klb)
		if err != nil {
			log.Error(err, "Could not generate cron cronJob")
			return ctrl.Result{}, err
		}

		log.Info("Creating a new BaseBackup resource")
		err = r.Create(ctx, dep)
		if err != nil && k8serrors.IsAlreadyExists(err) {
			log.Info("CronJob already exists")
		} else if err != nil {
			log.Error(err, "Failed to create new CronJob")
			return ctrl.Result{}, err
		} else {
			// cronJob created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		}
	}

	backupSchedule.InitFrom(cronJob)
	if !backupSchedule.IsEqual(klb) {
		backupSchedule.Update(klb)

		err = r.Update(ctx, backupSchedule.GetCronJob())
		if err != nil {
			log.Error(err, "Failed to update object")
			return ctrl.Result{}, err
		} else {
			log.Info("BaseBackup resource is updated")
		}
	} else {
		log.Info("No difference")
	}

	job, err := r.getBackupJob(ctx, klb)
	if err != nil {
		log.Error(err, "Failed to get the backup job")
		return ctrl.Result{}, err
	}
	// no backup jobs
	if job == nil {
		log.Info("No backup jobs found")
		klb.MarkUnknown("no backup jobs found")
		return ctrl.Result{}, nil
	}

	if running := backupSchedule.IsRunning(job); running {
		// notify kls that it has a running backup
		kl.BackupRunning(klb.Name)
		if err := r.Status().Update(ctx, kl); err != nil {
			return ctrl.Result{}, err
		}
		klb.MarkRunning(job.Name)
	} else {
		kl.BackupFinished()
		if err := r.Status().Update(ctx, kl); err != nil {
			return ctrl.Result{}, err
		}
		klb.MarkNotRunning()
	}

	if successful := backupSchedule.IsSuccessful(job); successful {
		klb.MarkSuccessful(job.Name)
	} else {
		klb.MarkFailed(job.Name)
	}
	if err := r.Status().Update(ctx, klb); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *KuberLogicBackupScheduleReconciler) cronJob(op interfaces.BackupSchedule, cmb *kuberlogicv1.KuberLogicBackupSchedule) (*v1beta1.CronJob, error) {
	op.Init(cmb)

	// Set kuberlogic backup instance as the owner and controller
	// if kuberlogic backup will remove -> dep also should be removed automatically
	err := ctrl.SetControllerReference(cmb, op.GetCronJob(), r.Scheme)
	if err != nil {
		return nil, err
	}

	return op.GetCronJob(), nil
}

func (r *KuberLogicBackupScheduleReconciler) getBackupJob(ctx context.Context, cmb *kuberlogicv1.KuberLogicBackupSchedule) (*v12.Job, error) {
	jobs := v12.JobList{}
	selector := &client.ListOptions{}

	client.InNamespace(cmb.Namespace).ApplyToList(selector)
	client.MatchingLabels{
		"backup-name": cmb.Name,
	}.ApplyToList(selector)

	if err := r.List(ctx, &jobs, selector); err != nil {
		return nil, err
	}
	if len(jobs.Items) < 1 {
		return nil, nil
	}

	sort.SliceStable(jobs.Items, func(i, j int) bool {
		first, second := jobs.Items[i], jobs.Items[j]
		return second.Status.StartTime.Before(first.Status.StartTime)
	})

	return &jobs.Items[0], nil
}

func (r *KuberLogicBackupScheduleReconciler) finalize(ctx context.Context, kl *kuberlogicv1.KuberLogicService, log logr.Logger) error {
	log.Info("finalizing backupschedule")

	kl.BackupFinished()
	return r.Status().Update(ctx, kl)
}

func (r *KuberLogicBackupScheduleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kuberlogicv1.KuberLogicBackupSchedule{}).
		Owns(&v1beta1.CronJob{}).
		Complete(r)
}
