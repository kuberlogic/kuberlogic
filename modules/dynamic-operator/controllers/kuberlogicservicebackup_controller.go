/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package controllers

import (
	"context"
	config "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/cfg"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/controllers/backuprestore"
	velero "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

// KuberlogicServiceBackupReconciler reconciles a KuberlogicServiceBackup object
type KuberlogicServiceBackupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cfg    *config.Config

	mu sync.Mutex
}

//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservicebackups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservicebackups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservicebackups/finalizers,verbs=update
//+kubebuilder:rbac:groups="velero.io",resources=backups;deletebackuprequests,verbs=list;watch
//+kubebuilder:rbac:groups="",resources=pods,verbs=list;watch

func (r *KuberlogicServiceBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx).WithValues("key", req.String(), "run", time.Now().UnixNano())

	l.Info("acquiring lock")
	r.mu.Lock()
	defer func() {
		r.mu.Unlock()
		l.Info("lock freed")
	}()

	klb := &kuberlogiccomv1alpha1.KuberlogicServiceBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Name,
		},
	}
	if err := r.Get(ctx, client.ObjectKeyFromObject(klb), klb); k8serrors.IsNotFound(err) {
		// not found, probably deleted
		return ctrl.Result{}, nil
	} else if err != nil {
		l.Error(err, "error getting object")
		return ctrl.Result{}, err
	}
	if klb.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(klb, backuprestore.BackupDeleteFinalizer) {
			controllerutil.AddFinalizer(klb, backuprestore.BackupDeleteFinalizer)
			return ctrl.Result{}, r.Update(ctx, klb)
		}
	}

	l = l.WithValues("phase", klb.Status.Phase)

	kls := &kuberlogiccomv1alpha1.KuberLogicService{}
	kls.SetName(klb.Spec.KuberlogicServiceName)

	if err := r.Get(ctx, client.ObjectKeyFromObject(kls), kls); k8serrors.IsNotFound(err) {
		l.Error(err, "service not found")
		return ctrl.Result{}, err
	} else if err != nil {
		l.Error(err, "error getting service", "name", kls.GetName())
		return ctrl.Result{}, err
	}

	// do not proceed if there is another backup / restore running
	if restoreRunning, restoreName := kls.RestoreRunning(); restoreRunning {
		l.Info("restore is running. will retry later", "restoreName", restoreName)
		return ctrl.Result{RequeueAfter: time.Second * 30}, nil
	}
	if backupRunning, backupName := kls.BackupRunning(); backupRunning && backupName != klb.GetName() {
		l.Info("another backup is running. will retry later", "backupName", backupName)
		return ctrl.Result{RequeueAfter: time.Second * 30}, nil
	}

	// sync current klb status to kls
	kls.SetBackupStatus(klb)
	if err := r.Status().Update(ctx, kls); err != nil {
		l.Error(err, "error syncing service backup status")
		return ctrl.Result{}, err
	}

	maxAttempts := 10
	if klb.Status.FailedAttempts >= maxAttempts {
		klb.MarkFailed("too many failures")
		return ctrl.Result{}, r.Status().Update(ctx, klb)
	}

	backup := backuprestore.NewVeleroBackupRestoreProvider(r.Client, l, kls, r.Cfg.Backups.SnapshotsEnabled)

	l.Info("syncing backup status")
	if err := backup.SetKuberlogicBackupStatus(ctx, klb); err != nil {
		l.Error(err, "error syncing backup status")
		return ctrl.Result{}, err
	}
	l.Info("backup status updated", "new status", klb.Status.Phase)

	if !klb.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(klb, backuprestore.BackupDeleteFinalizer) {
			if err := backup.BackupDeleteRequest(ctx, klb); err != nil {
				l.Error(err, "failed to delete backup")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		l.Info("object is being deleted but has no finalizers")
		return ctrl.Result{
			RequeueAfter: time.Second * 2,
		}, nil
	}

	var err error
	if klb.IsSuccessful() || klb.IsFailed() {
		if err = backup.AfterBackup(ctx, klb); err != nil {
			l.Error(err, "error during after backup routine")
		}
	} else if !klb.IsRequested() {
		if err = backup.BackupRequest(ctx, klb); err != nil {
			l.Error(err, "error planning backup")
		}
	}
	if err != nil {
		klb.IncreaseFailedAttemptCount()
		_ = r.Status().Update(ctx, klb)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KuberlogicServiceBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kuberlogiccomv1alpha1.KuberlogicServiceBackup{}).
		Owns(&velero.Backup{}).
		Owns(&velero.DeleteBackupRequest{}).
		Owns(&v1.Pod{}).
		Complete(r)
}
