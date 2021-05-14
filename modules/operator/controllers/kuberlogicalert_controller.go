package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/notifications"
	"github.com/kuberlogic/operator/modules/operator/util"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sync"
)

// KuberLogicAlertReconciler reconciles a KuberLogicAlert object
type KuberLogicAlertReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	mu     sync.Mutex

	NotificationsManager *notifications.NotificationManager
}

const (
	klaFinalizer = kuberlogicv1.Group + "/alert-finalizer"
)

// +kubebuilder:rbac:groups=cloudlinux.com,resources=kuberlogicalerts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudlinux.com,resources=kuberlogicalerts/status,verbs=get;update;patch
func (r *KuberLogicAlertReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("kuberlogicalert", req.NamespacedName)

	defer util.HandlePanic(log)

	mu := getMutex(req.NamespacedName)
	mu.Lock()
	defer mu.Unlock()

	log.Info("reconciliation started")
	// Fetch the KuberLogicAlert instance
	kla := &kuberlogicv1.KuberLogicAlert{}
	if err := r.Get(ctx, req.NamespacedName, kla); err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info(req.Namespace, req.Name, " has been deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get KuberLogicAlert")
		return ctrl.Result{}, err
	}
	// fetch the cluster information
	kls := &kuberlogicv1.KuberLogicService{}
	if err := r.Get(ctx, types.NamespacedName{Name: kla.Spec.Cluster, Namespace: req.Namespace}, kls); err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info(req.Namespace, kla.Spec.Cluster, " kuberlogicservice not found")
			return ctrl.Result{}, err
		}
		log.Error(err, "failed to get kuberlogicservice")
		return ctrl.Result{}, err
	}
	log = log.WithValues("kuberlogicservice", kls.Name)
	log.Info("kuberlogicservice is found")

	if kla.DeletionTimestamp != nil {
		log.Info("kuberlogicalert is pending for deletion")
		if controllerutil.ContainsFinalizer(kla, klaFinalizer) {
			if err := r.finalize(ctx, kla, kls, log); err != nil {
				log.Error(err, "error finalizing kuberlogicalert")
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(kla, klaFinalizer)
			if err := r.Update(ctx, kla); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	}

	if !controllerutil.ContainsFinalizer(kla, klaFinalizer) {
		log.Info("adding finalizer", "finalizer", klaFinalizer)
		controllerutil.AddFinalizer(kla, klaFinalizer)
		err := r.Update(ctx, kla)
		if err != nil {
			log.Error(err, "error adding finalizer")
		}
		return ctrl.Result{}, err
	}

	// check if we need to send a notification about new alert
	// only email notifications are supported
	if ignore := kla.IsSilenced() || kla.IsNotificationSent(); !ignore {
		if err := r.notifyNew(kla, kls.GetAlertEmail()); err != nil {
			log.Error(err, "notification sending failure")
			return ctrl.Result{}, err
		}

		log.Info("alert notification sent", "address", kls.GetAlertEmail())
		kla.NotificationSent(kls.GetAlertEmail())
		if err := r.Status().Update(ctx, kla); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		log.Info("notification conditions are not sent, skipping", "silenced", kla.IsSilenced(), "alreadySent", kla.IsNotificationSent())
	}
	return ctrl.Result{}, nil
}

// finalize function "resolves" an alert when kuberlogicalert is deleted.
func (r *KuberLogicAlertReconciler) finalize(ctx context.Context, kla *kuberlogicv1.KuberLogicAlert, kls *kuberlogicv1.KuberLogicService, log logr.Logger) error {
	log.Info("processing finalizer")
	err := r.notifyResolved(kla, kls.GetAlertEmail())
	log.Info("alert recovery notification sent", "address", kls.GetAlertEmail())
	return err
}

func (r *KuberLogicAlertReconciler) notifyNew(kla *kuberlogicv1.KuberLogicAlert, addr string) error {
	head := fmt.Sprintf("CRITICAL: SERVICE %s ALERT %s", kla.Spec.Cluster, kla.Spec.AlertName)

	if err := notifyEmail(addr, head, kla.Spec.Summary, r.NotificationsManager); err != nil {
		return err
	}
	return nil
}

func (r *KuberLogicAlertReconciler) notifyResolved(kla *kuberlogicv1.KuberLogicAlert, addr string) error {
	head := fmt.Sprintf("RESOLVED: SERVICE %s ALERT %s", kla.Spec.Cluster, kla.Spec.AlertName)
	message := fmt.Sprintf("Alert %s is now resolved.", kla.Spec.AlertName)

	if err := notifyEmail(addr, head, message, r.NotificationsManager); err != nil {
		return err
	}
	return nil
}

func (r *KuberLogicAlertReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kuberlogicv1.KuberLogicAlert{}).
		Complete(r)
}

func notifyEmail(addr, subj, body string, mgr *notifications.NotificationManager) error {
	ch, err := mgr.GetNotificationChannel(notifications.EmailChannel)
	if err != nil {
		return err
	}

	opts := map[string]string{
		"to": addr,
	}
	return ch.SendNotification(opts, subj, body)
}
