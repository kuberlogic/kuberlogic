package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/monitoring"
	"github.com/kuberlogic/operator/modules/operator/service-operator"
	v12 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

// KuberLogicBackupScheduleReconciler reconciles a KuberLogicBackupSchedule object
type KuberLogicBackupScheduleReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	mu     sync.Mutex
}

// +kubebuilder:rbac:groups=cloudlinux.com,resources=kuberlogicbackupschedules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudlinux.com,resources=kuberlogicbackupschedules/status,verbs=get;update;patch
func (r *KuberLogicBackupScheduleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("kuberlogicbackupschedule", req.NamespacedName)

	r.mu.Lock()
	defer r.mu.Unlock()

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

	op, err := service_operator.GetOperator(kl.Spec.Type)
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

	backupOperator, err := service_operator.GetBackupOperator(op)
	if err != nil {
		log.Error(err, "Could not define the backup operator")
		return ctrl.Result{}, err
	}
	backupOperator.SetBackupImage()
	backupOperator.SetBackupEnv(klb)

	cronJob := backupOperator.GetCronJob()
	err = r.Get(ctx,
		types.NamespacedName{
			Name:      klb.Name,
			Namespace: klb.Namespace,
		},
		cronJob)
	if err != nil && k8serrors.IsNotFound(err) {
		dep, err := r.cronJob(backupOperator, klb)
		if err != nil {
			log.Error(err, "Could not generate cron cronJob", "Name", klb.Name)
			return ctrl.Result{}, err
		}

		log.Info("Creating a new Backup resource", "Name", klb.Name)
		err = r.Create(ctx, dep)
		if err != nil && k8serrors.IsAlreadyExists(err) {
			log.Info("CronJob already exists", "Name", klb.Name)
		} else if err != nil {
			log.Error(err, "Failed to create new CronJob", "Name", klb.Name)
			return ctrl.Result{}, err
		} else {
			// cronJob created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		}
	}

	backupOperator.InitFrom(cronJob)
	if !backupOperator.IsEqual(klb) {
		backupOperator.Update(klb)

		err = r.Update(ctx, backupOperator.GetCronJob())
		if err != nil {
			log.Error(err, "Failed to update object", "Name", klb.Name)
			return ctrl.Result{}, err
		} else {
			log.Info("Backup resource is updated", "Name", klb.Name)
		}
	} else {
		log.Info("No difference", "Name", klb.Name)
	}

	jobList, err := r.getJobList(ctx, klb)
	if err != nil {
		log.Error(err, "Failed to receive list of jobs",
			"Name", klb.Name)
		return ctrl.Result{}, err
	}

	status := backupOperator.CurrentStatus(jobList)
	if !klb.IsEqual(status) {
		klb.SetStatus(status)
		err = r.Update(ctx, klb)
		//err = r.Status().Update(ctx, kl) # FIXME: Figure out why it's failed
		if err != nil {
			log.Error(err, "Failed to update kl backup object")
		} else {
			log.Info("KuberLogicBackupSchedule status is updated", "Status", klb.GetStatus())
		}
	}

	monitoring.KuberLogicBackupSchedules[monitoringKey] = klb

	return ctrl.Result{}, nil
}

func (r *KuberLogicBackupScheduleReconciler) cronJob(op service_operator.Backup, cmb *kuberlogicv1.KuberLogicBackupSchedule) (*v1beta1.CronJob, error) {
	op.Init(cmb)

	// Set kuberlogic backup instance as the owner and controller
	// if kuberlogic backup will remove -> dep also should be removed automatically
	err := ctrl.SetControllerReference(cmb, op.GetCronJob(), r.Scheme)
	if err != nil {
		return nil, err
	}

	return op.GetCronJob(), nil
}

func (r *KuberLogicBackupScheduleReconciler) getJobList(ctx context.Context, cmb *kuberlogicv1.KuberLogicBackupSchedule) (v12.JobList, error) {
	jobs := v12.JobList{}
	selector := &client.ListOptions{}

	client.InNamespace(cmb.Namespace).ApplyToList(selector)
	client.MatchingLabels{
		"backup-name": cmb.Name,
	}.ApplyToList(selector)

	err := r.List(ctx, &jobs, selector)
	return jobs, err
}

func (r *KuberLogicBackupScheduleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kuberlogicv1.KuberLogicBackupSchedule{}).
		Owns(&v1beta1.CronJob{}).
		Complete(r)
}
