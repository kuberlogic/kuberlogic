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

var (
	backupRestoreRequeueAfter = time.Minute * 1
	notReadyBeforeFailed      = time.Minute * 5
)

//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservices/finalizers,verbs=update

// compose plugin roles:
//+kubebuilder:rbac:groups="",resources=serviceaccounts;services;persistentvolumeclaims;,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments;,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

func (r *KuberLogicServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logger.FromContext(ctx).WithValues("kuberlogicservicetype", req.String(), "run", time.Now().UnixNano())
	log.Info("Reconciliation started")
	defer HandlePanic(log)

	r.mu.Lock()
	defer r.mu.Unlock()

	// Fetch the KuberLogicServices instance
	kls := &kuberlogiccomv1alpha1.KuberLogicService{}
	err := r.Get(ctx, req.NamespacedName, kls)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("object not found", "key", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get KuberLogicService")
		return ctrl.Result{}, err
	}

	if backupRunning, backupName := kls.BackupRunning(); backupRunning {
		klb := &kuberlogiccomv1alpha1.KuberlogicServiceBackup{}
		klb.SetName(backupName)
		if err := r.Get(ctx, client.ObjectKeyFromObject(klb), klb); err != nil {
			if k8serrors.IsNotFound(err) {
				log.Info("backup request is not found")
				kls.SetBackupStatus(nil)
				return ctrl.Result{}, r.Status().Update(ctx, kls)
			}
			log.Error(err, "error getting backup object")
			return ctrl.Result{}, err
		}

		// requeue
		return ctrl.Result{RequeueAfter: backupRestoreRequeueAfter}, nil
	}
	if restoreRunning, restoreName := kls.RestoreRunning(); restoreRunning {
		klr := &kuberlogiccomv1alpha1.KuberlogicServiceRestore{}
		klr.SetName(restoreName)
		if err := r.Get(ctx, client.ObjectKeyFromObject(klr), klr); err != nil {
			if k8serrors.IsNotFound(err) {
				log.Info("restore request is not found")
				kls.SetRestoreStatus(nil)
				return ctrl.Result{}, r.Status().Update(ctx, kls)
			}
			log.Error(err, "error getting restore object")
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: backupRestoreRequeueAfter}, nil
	}

	env := kuberlogicserviceenv.New(r.Client, kls, r.Cfg)
	if err := env.SetupEnv(ctx); err != nil {
		kls.ConfigurationFailed("KuberlogicService environment")
		_ = r.Status().Update(ctx, kls)

		log.Error(err, "error setting up KuberlogicService environment")
		return ctrl.Result{}, err
	}
	ns := env.NamespaceName

	spec := make(map[string]interface{}, 0)
	if len(kls.Spec.Advanced.Raw) > 0 {
		if err := json.Unmarshal(kls.Spec.Advanced.Raw, &spec); err != nil {
			kls.ConfigurationFailed("KuberlogicService environment")
			_ = r.Status().Update(ctx, kls)

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

		kls.ConfigurationFailed(pluginLoadedErr.Error())
		_ = r.Status().Update(ctx, kls)

		return ctrl.Result{}, pluginLoadedErr
	}

	pluginRequest := commons.PluginRequest{
		Name:          kls.Name,
		Namespace:     ns,
		Replicas:      kls.Spec.Replicas,
		VolumeSize:    kls.Spec.VolumeSize,
		Version:       kls.Spec.Version,
		TLSEnabled:    kls.TLSEnabled(),
		TLSSecretName: r.Cfg.SvcOpts.TLSSecretName,
		Host:          kls.GetHost(),
		Parameters:    spec,
	}

	if err := pluginRequest.SetLimits(&kls.Spec.Limits); err != nil {
		kls.ConfigurationFailed("plugin error: " + err.Error())
		_ = r.Status().Update(ctx, kls)

		log.Error(err, "error converting resources")
		return ctrl.Result{}, err
	}
	resp := plugin.Convert(pluginRequest)
	if resp.Error() != nil {
		kls.ConfigurationFailed("plugin error (Convert): " + resp.Error().Error())
		_ = r.Status().Update(ctx, kls)

		log.Error(resp.Error(), "error from rpc call 'Convert'", "plugin request", pluginRequest)
		return ctrl.Result{}, resp.Error()
	}
	log.Info("=========", "plugin response", resp, "plugin request", pluginRequest)

	// collect cluster objects
	for _, o := range resp.Objects {
		o.SetNamespace(env.NamespaceName)
		if err := r.Get(ctx, client.ObjectKeyFromObject(o), o); k8serrors.IsNotFound(err) {
			// object not found continue iterating
			continue
		} else if err != nil {
			kls.ClusterSyncFailed(fmt.Sprintf("failed to syc %s %s/%s", o.GetKind(), o.GetNamespace(), o.GetName()))
			_ = r.Status().Update(ctx, kls)

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
		kls.ConfigurationFailed("plugin error (Convert): " + resp.Error().Error())
		_ = r.Status().Update(ctx, kls)

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
			kls.ClusterSyncFailed(fmt.Sprintf("failed to syc %s %s/%s", o.GetKind(), o.GetNamespace(), o.GetName()))
			_ = r.Status().Update(ctx, kls)

			log.Error(err, "error syncing object", "object", o)
			return ctrl.Result{}, err
		}
		log.Info("synced object", "op", op, "object", o)
	}

	// pause service when requested
	if kls.PauseRequested() {
		if err := env.PauseService(ctx); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "error pausing service")
		}
		kls.MarkPaused()
	} else if kls.Resumed() {
		if err := env.ResumeService(ctx); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "error resuming service")
		}
		kls.MarkResumed()
	}

	// expose service
	kls.SetAccessEndpoint()

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
		kls.ConfigurationFailed("plugin error (Status): " + resp.Error().Error())
		_ = r.Status().Update(ctx, kls)

		log.Error(resp.Error(), "error from rpc call 'ForUpdate'")
		return ctrl.Result{}, resp.Error()
	}

	var requeueAfter time.Duration
	if status.IsReady {
		kls.MarkReady("ReadyConditionMet")
	} else {
		klsReady, _, transitionTime := kls.IsReady()
		if !klsReady && transitionTime != nil && time.Since(*transitionTime) > notReadyBeforeFailed {
			kls.ClusterSyncFailed("service is not ready for too long")
			requeueAfter = time.Minute * 5
		} else {
			kls.MarkNotReady("ReadyConditionNotMet")
		}
	}

	if err := r.Status().Update(ctx, kls); err != nil {
		log.Error(err, "error syncing status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

func (r *KuberLogicServiceReconciler) SetupWithManager(mgr ctrl.Manager, objects ...client.Object) error {
	builder := ctrl.NewControllerManagedBy(mgr).For(
		&kuberlogiccomv1alpha1.KuberLogicService{})

	for _, object := range objects {
		builder = builder.Owns(object)
	}

	builder.Owns(&v1.Namespace{})
	builder.Owns(&v12.NetworkPolicy{})
	builder.Owns(&v1.Secret{})
	builder.Owns(&kuberlogiccomv1alpha1.KuberlogicServiceBackupSchedule{})
	return builder.Complete(r)
}
