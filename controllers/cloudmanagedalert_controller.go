package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

// CloudManagedAlertReconciler reconciles a CloudManagedAlert object
type CloudManagedAlertReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	mu     sync.Mutex
}

// +kubebuilder:rbac:groups=cloudlinux.com,resources=cloudmanagedalerts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudlinux.com,resources=cloudmanagedalerts/status,verbs=get;update;patch
func (r *CloudManagedAlertReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("cloudmanagedalert", req.NamespacedName)

	r.mu.Lock()
	defer r.mu.Unlock()

	// Fetch the CloudManagedAlert instance
	cloudmanagedAlert := &cloudlinuxv1.CloudManagedAlert{}
	err := r.Get(ctx, req.NamespacedName, cloudmanagedAlert)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info(req.Namespace, req.Name, " has been deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get CloudManagedAlert")
		return ctrl.Result{}, err
	}

	// Trigger alert processing and update status if needed.
	// This will trigger reconcilation loop one more time.
	newStatus := r.Process(cloudmanagedAlert)
	if !cloudmanagedAlert.IsEqual(newStatus) {
		cloudmanagedAlert.SetStatus(newStatus)

		err := r.Update(ctx, cloudmanagedAlert)
		if err != nil {
			log.Error(err, "Failed to update alert status")
			return ctrl.Result{}, err
		} else {
			log.Info("Alert status is updated", "Status", cloudmanagedAlert.Status.Status)
		}
	}

	return ctrl.Result{}, nil
}

func (r *CloudManagedAlertReconciler) Process(cla *cloudlinuxv1.CloudManagedAlert) (status string) {
	log := r.Log.WithValues("cloudmanagedalert", cla.Name)
	log.Info("Running alert processing")

	switch cla.Status.Status {
	case "":
		return NewAlertProcess(r, cla, log)
	case cloudlinuxv1.AlertCreatedStatus:
		return CreatedAlertProcess(r, cla, log)
	case cloudlinuxv1.AlertAckedStatus:
		return AckedAlertProcess(r, cla, log)
	case cloudlinuxv1.AlertResolvedStatus:
		return cloudlinuxv1.AlertResolvedStatus
	default:
		return cloudlinuxv1.AlertUnknownStatus
	}
}

func (r *CloudManagedAlertReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudlinuxv1.CloudManagedAlert{}).
		Complete(r)
}

var AlertsDB = map[string]func(r *CloudManagedAlertReconciler, cla *cloudlinuxv1.CloudManagedAlert) error{
	"cloudmanaged-memory-usage": func(r *CloudManagedAlertReconciler, cla *cloudlinuxv1.CloudManagedAlert) error {
		// just increases memory limits

		log := r.Log.WithValues("cloudmanagedalert", cla.Name)

		cm := &cloudlinuxv1.CloudManaged{}
		if err := r.Get(context.Background(), types.NamespacedName{
			Name:      cla.Spec.Cluster,
			Namespace: cla.Namespace,
		}, cm); err != nil {
			log.Error(err, "Error getting related cloudmanaged cluster")
			return err
		}

		// calculate new memory limits
		// this is just an example so implementation is very dumb
		cm.Spec.Resources.Limits = v1.ResourceList{
			v1.ResourceCPU:    cm.Spec.Resources.Limits.Cpu().DeepCopy(),
			v1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dKi", cm.Spec.Resources.Limits.Memory().Value()*2/1024)),
		}
		if err := r.Update(context.Background(), cm); err != nil {
			log.Error(err, "Error updating resources for related CloudManaged object!")
			return err
		}
		return nil
	},
}

func NewAlertProcess(r *CloudManagedAlertReconciler, cla *cloudlinuxv1.CloudManagedAlert, log logr.Logger) (status string) {
	log.Info("Received an alert")
	return cloudlinuxv1.AlertCreatedStatus
}

func CreatedAlertProcess(r *CloudManagedAlertReconciler, cla *cloudlinuxv1.CloudManagedAlert, log logr.Logger) (status string) {
	log.Info("Alert is valid and is ready to be processed")
	return cloudlinuxv1.AlertAckedStatus
}

func AckedAlertProcess(r *CloudManagedAlertReconciler, cla *cloudlinuxv1.CloudManagedAlert, log logr.Logger) (status string) {
	log.Info("Processing alert")

	alertAction := AlertsDB[cla.Spec.AlertName]
	if alertAction == nil {
		log.Info("No meaningful action found", "alert type", cla.Spec.AlertName)
		return cloudlinuxv1.AlertUnknownStatus
	} else {
		log.Info("Found an action for alert")
	}
	if err := alertAction(r, cla); err != nil {
		log.Error(err, "Error processing action for alert", "alert type", cla.Spec.AlertName)
		return cloudlinuxv1.AlertUnknownStatus
	}

	return cloudlinuxv1.AlertResolvedStatus
}
