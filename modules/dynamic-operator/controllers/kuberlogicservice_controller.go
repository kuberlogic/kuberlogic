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
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/cfg"
	kuberlogicserviceenv "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/controllers/kuberlogicservice-env"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/api/networking/v1"
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

	Cfg *cfg.Config
	mu  sync.Mutex
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

// compose plugin roles:
//+kubebuilder:rbac:groups="",resources=serviceaccounts;services;persistentvolumeclaims;,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments;,verbs=get;list;watch;create;update;patch;delete

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
	env, err := kuberlogicserviceenv.SetupEnv(kls, r.Client, r.Cfg, ctx)
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
		TLSEnabled: kls.TLSEnabled(),
		Host:       kls.GetHost(),
		Parameters: spec,
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
		o.SetNamespace(env.NamespaceName)
		if err := r.Get(ctx, client.ObjectKeyFromObject(o), o); k8serrors.IsNotFound(err) {
			// object not found conitnue iterating
			continue
		} else if err != nil {
			log.Error(err, "error fetching object from cluster", "object", o)
			return ctrl.Result{}, err
		} else {
			// object found
			pluginRequest.AddObject(o)
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
		o.SetNamespace(env.NamespaceName)
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

	// pause service when requested
	if kls.Paused() {
		if err := env.PauseService(); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "error pausing service")
		}
	}

	// expose service
	endpoint, err := env.ExposeService(resp.Service, resp.Protocol == commons.HTTPproto && kls.GetHost() != "")
	if err != nil {
		log.Error(err, "error exposing service")
		return ctrl.Result{}, err
	}
	kls.SetAccessEndpoint(endpoint)

	// sync status
	log.Info("syncing status")
	statusRequest := &commons.PluginRequest{
		Name:       kls.Name,
		Namespace:  kls.Namespace,
		Parameters: spec,
	}
	statusRequest.SetObjects(resp.Objects)
	status := plugin.Status(*statusRequest)
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

	builder.Owns(&v1.Namespace{})
	builder.Owns(&v12.NetworkPolicy{})
	builder.Owns(&v12.Ingress{})
	builder.Owns(&v1.ResourceQuota{})
	builder.Owns(&v1.Secret{})
	return builder.Complete(r)
}
