/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package controllers

import (
	"context"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/cfg"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/controllers/backuprestore"
	"github.com/pkg/errors"
	velero "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

// KuberlogicServiceRestoreReconciler reconciles a KuberlogicServiceRestore object
type KuberlogicServiceRestoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cfg    *cfg.Config

	mu sync.Mutex
}

//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservicerestores,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservicerestores/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservicerestores/finalizers,verbs=update
//+kubebuilder:rbac:groups="velero.io",resources=restores,verbs=list;watch

func (r *KuberlogicServiceRestoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx).WithValues("key", req.String(), "run", time.Now().UnixNano())

	l.Info("acquiring lock")
	r.mu.Lock()
	defer func() {
		r.mu.Unlock()
		l.Info("lock freed")
	}()

	klr := &kuberlogiccomv1alpha1.KuberlogicServiceRestore{}
	if err := r.Get(ctx, req.NamespacedName, klr); err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("object not found")
			return ctrl.Result{}, nil
		}
		l.Error(err, "failed to get klr")
		return ctrl.Result{}, err
	}

	klb := &kuberlogiccomv1alpha1.KuberlogicServiceBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name: klr.Spec.KuberlogicServiceBackup,
		},
	}
	if err := r.Get(ctx, client.ObjectKeyFromObject(klb), klb); err != nil {
		// not found, probably deleted
		l.Error(err, "failed to get backup for the restore", "backup", klb.GetName())
		return ctrl.Result{}, err
	}
	if !klb.IsSuccessful() {
		err := errors.New("backup must be successful")
		l.Error(err, "source backup is not successful", "backup", klb)
		return ctrl.Result{}, err
	}

	kls := &kuberlogiccomv1alpha1.KuberLogicService{}
	kls.SetName(klb.Spec.KuberlogicServiceName)

	if err := r.Get(ctx, client.ObjectKeyFromObject(kls), kls); k8serrors.IsNotFound(err) {
		l.Error(err, "service not found")
		return ctrl.Result{}, err
	} else if err != nil {
		l.Error(err, "failed to get service for restore", "name", kls.GetName())
		return ctrl.Result{}, err
	}

	l = l.WithValues("phase", klr.Status.Phase)

	// do not proceed if there is another backup / restore running
	if restoreRunning, restoreName := kls.RestoreRunning(); restoreRunning && restoreName != klr.GetName() {
		l.Info("restore is running. will retry later", "restoreName", restoreName)
		return ctrl.Result{RequeueAfter: time.Second * 30}, nil
	}
	if backupRunning, backupName := kls.BackupRunning(); backupRunning {
		l.Info("another backup is running. will retry later", "backupName", backupName)
		return ctrl.Result{RequeueAfter: time.Second * 30}, nil
	}

	// mark service as being restored
	kls.SetRestoreStatus(klr)
	if err := r.Status().Update(ctx, kls); err != nil {
		l.Error(err, "failed to mark kls restore status")
		return ctrl.Result{}, err
	}

	maxAttempts := 10
	if klr.Status.FailedAttempts >= maxAttempts {
		klr.MarkFailed("too many failures")
		return ctrl.Result{}, r.Status().Update(ctx, klr)
	}

	restore := backuprestore.NewVeleroBackupRestoreProvider(r.Client, l, kls, r.Cfg.Backups.SnapshotsEnabled)

	l.Info("syncing restore status")
	if err := restore.SetKuberlogicRestoreStatus(ctx, klr); err != nil {
		l.Error(err, "error syncing backup status")
		return ctrl.Result{}, err
	}
	l.Info("restore status updated", "new status", klr.Status.Phase)

	var err error
	if klr.IsFailed() || klr.IsSuccessful() {
		if err = restore.AfterRestore(ctx, klr); err != nil {
			l.Error(err, "failed to execute after restore routine")
		}
	} else if !klr.IsRequested() {
		if err = restore.RestoreRequest(ctx, klb, klr); err != nil {
			l.Error(err, "failed to start restore")
		}
	}
	if err != nil {
		klr.IncreaseFailedAttemptCount()
		_ = r.Status().Update(ctx, klr)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KuberlogicServiceRestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kuberlogiccomv1alpha1.KuberlogicServiceRestore{}).
		Owns(&velero.Restore{}).
		Complete(r)
}
