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
	kuberlogicserviceenv "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/controllers/kuberlogicservice-env"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	"sync"
	"time"
)

// KuberLogicServiceReconciler reconciles a KuberLogicService object
type KuberLogicServiceReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Plugins map[string]commons.PluginService

	mu sync.Mutex
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

	r.mu.Lock()
	defer r.mu.Unlock()

	// Fetch the KuberLogicServices instance
	kls := &kuberlogiccomv1alpha1.KuberLogicService{}
	err := r.Get(ctx, req.NamespacedName, kls)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("object not found", "key", req.NamespacedName)

			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get KuberLogicService")
		return ctrl.Result{}, err
	}

	log.Info("verifying KuberLogicService environment")
	env, err := kuberlogicserviceenv.SetupEnv(kls, r.Client, ctx)
	if err != nil {
		log.Error(err, "error setting up KuberlogicService environment")
		return ctrl.Result{}, err
	}
	ns := env.NamespaceName

	spec := make(map[string]interface{}, 0)
	if len(kls.Spec.Advanced.Raw) > 0 {
		if err := json.Unmarshal(kls.Spec.Advanced.Raw, &spec); err != nil {
			log.Error(err, "error unmarshalling spec")
			return ctrl.Result{}, err
		}
	}

	log.Info("spec", "spec", kls.Spec)
	log.Info("plugin type", "type", kls.Spec.Type)
	log = log.WithValues("plugin", kls.Spec.Type)

	plugin, found := r.Plugins[kls.Spec.Type]
	if !found {
		pluginLoadedErr := errors.New("plugin not found")
		log.Error(pluginLoadedErr, "")
		return ctrl.Result{}, errors.New("plugin not found")
	}

	pluginRequest := commons.PluginRequest{
		Name:       kls.Name,
		Namespace:  ns,
		Replicas:   kls.Spec.Replicas,
		VolumeSize: kls.Spec.VolumeSize,
		Version:    kls.Spec.Version,
		Parameters: spec,
		Objects:    nil,
	}

	if kls.Spec.Domain != "" {
		pluginRequest.Host = kls.Name + "." + kls.Spec.Domain
	}

	if err := pluginRequest.SetLimits(&kls.Spec.Limits); err != nil {
		log.Error(err, "error converting resources")
		return ctrl.Result{}, err
	}
	resp := plugin.Convert(pluginRequest)
	if resp.Error() != nil {
		log.Error(resp.Error(), "error from rpc call 'Convert'", "plugin request", pluginRequest)
		return ctrl.Result{}, resp.Error()
	}
	log.Info("=========", "plugin response", resp, "plugin request", pluginRequest)

	// collect cluster objects
	for _, o := range resp.Objects {
		if err := r.Get(ctx, client.ObjectKeyFromObject(o), o); k8serrors.IsNotFound(err) {
			// object not found conitnue iterating
			continue
		} else if err != nil {
			log.Error(err, "error fetching object from cluster", "object", o)
			return ctrl.Result{}, err
		} else {
			// object found
			pluginRequest.Objects = append(pluginRequest.Objects, o)
		}
	}

	// convert found objects
	resp = plugin.Convert(pluginRequest)
	if resp.Error() != nil {
		log.Error(resp.Error(), "error from rpc call 'Convert'", "plugin request", pluginRequest)
		return ctrl.Result{}, resp.Error()
	}
	log.Info("=========", "plugin response", resp, "plugin request", pluginRequest)

	// now create or update objects in cluster
	for _, o := range resp.Objects {
		desiredState := o.UnstructuredContent()
		op, err := ctrl.CreateOrUpdate(ctx, r.Client, o, func() error {
			log.Info("mutating object", "current", o.UnstructuredContent(), "desired", desiredState)
			o.SetUnstructuredContent(desiredState)
			return ctrl.SetControllerReference(kls, o, r.Scheme)
		})
		if err != nil {
			log.Error(err, "error syncing object", "object", o)
			return ctrl.Result{}, err
		}
		log.Info("synced object", "op", op, "object", o)
	}

	// sync status
	log.Info("syncing status")
	status := plugin.Status(commons.PluginRequest{
		Name:       kls.Name,
		Namespace:  ns,
		Objects:    resp.Objects,
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
