/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/go-logr/logr"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// KuberLogicServiceReconciler reconciles a KuberLogicService object
type KuberLogicServiceReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Plugins map[string]commons.PluginService
}

func HandlePanic(log logr.Logger) {
	if err := recover(); err != nil {
		log.Error(errors.New("handle panic"), fmt.Sprintf("%v", err))
		result := sentry.Flush(5 * time.Second)
		if !result {
			time.Sleep(5 * time.Second)
		}
		panic(err)
	}
}

//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservices/finalizers,verbs=update

func (r *KuberLogicServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logger.FromContext(ctx).WithValues("kuberlogicservicetype", req.String())
	log.Info("Reconciliation started")
	defer HandlePanic(log)

	// Fetch the KuberLogicServices instance
	kls := &kuberlogiccomv1alpha1.KuberLogicService{}
	err := r.Get(ctx, req.NamespacedName, kls)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info(req.Namespace, req.Name, "is absent")

			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get KuberLogicService")
		return ctrl.Result{}, err
	}

	spec := make(map[string]interface{}, 0)
	if len(kls.Spec.Advanced.Raw) > 0 {
		if err := json.Unmarshal(kls.Spec.Advanced.Raw, &spec); err != nil {
			log.Error(err, "error unmarshalling spec")
			return ctrl.Result{}, err
		}
	}

	log.Info("spec", "spec", kls.Spec)
	log.Info("plugin type", "type", kls.Spec.Type)

	//return ctrl.Result{}, nil

	plugin := r.Plugins[kls.Spec.Type]
	resp := plugin.Type()
	if resp.Error() != nil {
		log.Error(resp.Error(), "error from rpc call 'Empty'")
		return ctrl.Result{}, resp.Error()
	}

	svc := resp.Object
	svc.SetName(kls.Name)
	svc.SetNamespace(kls.Namespace)
	if err := r.Client.Get(ctx, req.NamespacedName, svc); k8serrors.IsNotFound(err) {
		log.Info("creating new service", "type", kls.Spec.Type)

		req := commons.PluginRequest{
			Name:       kls.Name,
			Namespace:  kls.Namespace,
			Replicas:   kls.Spec.Replicas,
			VolumeSize: kls.Spec.VolumeSize,
			Version:    kls.Spec.Version,
			Parameters: spec,
		}
		err = req.SetLimits(&kls.Spec.Limits)
		if err != nil {
			log.Error(err, "error from converting resources")
			return ctrl.Result{}, err
		}

		resp := plugin.Convert(req)
		if resp.Error() != nil {
			log.Error(resp.Error(), "error from rpc call 'ForCreate'")
			return ctrl.Result{}, resp.Error()
		}
		log.Info("creating service", "object", resp)
		svc := resp.Object

		if err := ctrl.SetControllerReference(kls, svc, r.Scheme); err != nil {
			log.Error(err, "error setting controller reference")
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, svc); err != nil {
			log.Error(err, "error creating service object")
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "failed to get service object")
		return ctrl.Result{}, err
	} else {

		req := commons.PluginRequest{
			Name:       kls.Name,
			Namespace:  kls.Namespace,
			Object:     svc,
			Replicas:   kls.Spec.Replicas,
			VolumeSize: kls.Spec.VolumeSize,
			Version:    kls.Spec.Version,
			Parameters: spec,
		}
		err = req.SetLimits(&kls.Spec.Limits)
		if err != nil {
			log.Error(err, "error from converting resources")
			return ctrl.Result{}, err
		}
		resp = plugin.Convert(req)
		if resp.Error() != nil {
			log.Error(resp.Error(), "error from rpc call 'ForUpdate'")
			return ctrl.Result{}, resp.Error()
		}

		log.Info("updating service", "object", resp)
		svc = resp.Object

		if err := r.Update(ctx, svc); err != nil {
			log.Error(err, "error updating service object")
			return ctrl.Result{}, err
		}
	}

	log.Info("syncing status", "object", svc.UnstructuredContent())
	status := plugin.Status(commons.PluginRequest{
		Name:       kls.Name,
		Namespace:  kls.Namespace,
		Object:     svc,
		Parameters: spec,
	})
	if resp.Error() != nil {
		log.Error(resp.Error(), "error from rpc call 'ForUpdate'")
		return ctrl.Result{}, resp.Error()
	}
	if status.IsReady {
		kls.MarkReady("ReadyConditionMet")
	} else {
		kls.MarkNotReady("ReadyConditionNotMet")
	}

	if err := r.Status().Update(ctx, kls); err != nil {
		log.Error(err, "error syncing status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *KuberLogicServiceReconciler) SetupWithManager(mgr ctrl.Manager, objects ...client.Object) error {
	builder := ctrl.NewControllerManagedBy(mgr).For(
		&kuberlogiccomv1alpha1.KuberLogicService{})

	for _, object := range objects {
		builder = builder.Owns(object)
	}
	return builder.Complete(r)
}
