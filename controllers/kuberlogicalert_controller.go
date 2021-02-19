package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	kuberlogicv1 "gitlab.com/cloudmanaged/operator/api/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

// KuberLogicAlertReconciler reconciles a KuberLogicAlert object
type KuberLogicAlertReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	mu     sync.Mutex
}

// +kubebuilder:rbac:groups=cloudlinux.com,resources=kuberlogicalerts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudlinux.com,resources=kuberlogicalerts/status,verbs=get;update;patch
func (r *KuberLogicAlertReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("kuberlogicalert", req.NamespacedName)

	r.mu.Lock()
	defer r.mu.Unlock()

	// Fetch the KuberLogicAlert instance
	klAlert := &kuberlogicv1.KuberLogicAlert{}
	err := r.Get(ctx, req.NamespacedName, klAlert)
	if err != nil {
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

	// Trigger alert processing and update status if needed.
	// This will trigger reconcilation loop one more time.
	newStatus := r.Process(klAlert)
	if !klAlert.IsEqual(newStatus) {
		klAlert.SetStatus(newStatus)

		err := r.Update(ctx, klAlert)
		if err != nil {
			log.Error(err, "Failed to update alert status")
			return ctrl.Result{}, err
		} else {
			log.Info("Alert status is updated", "Status", klAlert.Status.Status)
		}
	}

	return ctrl.Result{}, nil
}

func (r *KuberLogicAlertReconciler) Process(cla *kuberlogicv1.KuberLogicAlert) (status string) {
	log := r.Log.WithValues("kuberlogicalert", cla.Name)
	log.Info("Running alert processing")

	switch cla.Status.Status {
	case "":
		return NewAlertProcess(r, cla, log)
	case kuberlogicv1.AlertCreatedStatus:
		return CreatedAlertProcess(r, cla, log)
	case kuberlogicv1.AlertAckedStatus:
		return AckedAlertProcess(r, cla, log)
	case kuberlogicv1.AlertResolvedStatus:
		return kuberlogicv1.AlertResolvedStatus
	default:
		return kuberlogicv1.AlertUnknownStatus
	}
}

func (r *KuberLogicAlertReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kuberlogicv1.KuberLogicAlert{}).
		Complete(r)
}

var AlertsDB = map[string]func(r *KuberLogicAlertReconciler, cla *kuberlogicv1.KuberLogicAlert) error{
	"kuberlogic-memory-usage": func(r *KuberLogicAlertReconciler, cla *kuberlogicv1.KuberLogicAlert) error {
		// just increases memory limits

		log := r.Log.WithValues("kuberlogicalert", cla.Name)

		cm := &kuberlogicv1.KuberLogicService{}
		if err := r.Get(context.Background(), types.NamespacedName{
			Name:      cla.Spec.Cluster,
			Namespace: cla.Namespace,
		}, cm); err != nil {
			log.Error(err, "Error getting related kuberlogic cluster")
			return err
		}

		// calculate new memory limits
		// this is just an example so implementation is very dumb
		cm.Spec.Resources.Limits = v1.ResourceList{
			v1.ResourceCPU:    cm.Spec.Resources.Limits.Cpu().DeepCopy(),
			v1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dKi", cm.Spec.Resources.Limits.Memory().Value()*2/1024)),
		}
		if err := r.Update(context.Background(), cm); err != nil {
			log.Error(err, "Error updating resources for related KuberLogicService object!")
			return err
		}
		return nil
	},
}

func NewAlertProcess(r *KuberLogicAlertReconciler, cla *kuberlogicv1.KuberLogicAlert, log logr.Logger) (status string) {
	log.Info("Received an alert")
	return kuberlogicv1.AlertCreatedStatus
}

func CreatedAlertProcess(r *KuberLogicAlertReconciler, cla *kuberlogicv1.KuberLogicAlert, log logr.Logger) (status string) {
	log.Info("Alert is valid and is ready to be processed")
	return kuberlogicv1.AlertAckedStatus
}

func AckedAlertProcess(r *KuberLogicAlertReconciler, cla *kuberlogicv1.KuberLogicAlert, log logr.Logger) (status string) {
	log.Info("Processing alert")

	alertAction := AlertsDB[cla.Spec.AlertName]
	if alertAction == nil {
		log.Info("No meaningful action found", "alert type", cla.Spec.AlertName)
		return kuberlogicv1.AlertUnknownStatus
	} else {
		log.Info("Found an action for alert")
	}
	if err := alertAction(r, cla); err != nil {
		log.Error(err, "Error processing action for alert", "alert type", cla.Spec.AlertName)
		return kuberlogicv1.AlertUnknownStatus
	}

	return kuberlogicv1.AlertResolvedStatus
}
