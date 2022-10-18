/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/cfg"
	kuberlogicserviceenv "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/controllers/kuberlogicservice-env"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	"github.com/pkg/errors"
	"io"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	"sync"
	"time"
)

var NewRemoteExecutor = remotecommand.NewSPDYExecutor

// KuberLogicServiceReconciler reconciles a KuberLogicService object
type KuberLogicServiceReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Plugins map[string]commons.PluginService

	Cfg        *cfg.Config
	RESTConfig *rest.Config

	mu sync.Mutex
}

func HandlePanic() {
	if err := recover(); err != nil {
		sentry.CaptureException(errors.New(fmt.Sprintf("panic captured: %v", err)))
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
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods/exec,verbs=get;list;watch;create;update;patch;delete

// compose plugin roles:
//+kubebuilder:rbac:groups="",resources=serviceaccounts;services;persistentvolumeclaims;secrets;configmaps;,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments;,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

func (r *KuberLogicServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logger.FromContext(ctx).WithValues("kuberlogicservicetype", req.String(), "run", time.Now().UnixNano())
	log.Info("Reconciliation started")
	defer HandlePanic()

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

	if kls.Archived() {
		// exit from reconciliation
		log.Info("service is archived")
		return ctrl.Result{}, nil
	}

	env := kuberlogicserviceenv.New(r.Client, kls, r.Cfg)

	// archive service when requested
	if kls.ArchiveRequested() {
		log.Info("request to archive service")
		if err := env.ArchiveService(ctx); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "error archived service")
		}
		log.Info("service archived successfully")
		kls.MarkArchived()
		return ctrl.Result{}, r.Status().Update(ctx, kls)
	} else if kls.UnarchiveRequested() {
		log.Info("service unarchived successfully")
		kls.MarkUnarchived()
		return ctrl.Result{}, r.Status().Update(ctx, kls)
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
		Version:       kls.Spec.Version,
		Insecure:      kls.Insecure(),
		TLSSecretName: r.Cfg.SvcOpts.TLSSecretName,
		Host:          kls.GetHost(),
		Parameters:    spec,

		IngressClass: r.Cfg.IngressClass,
		StorageClass: r.Cfg.StorageClass,
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
	log.Info("DEBUG", "plugin response", resp, "plugin request", pluginRequest)

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

	// set service access endpoint
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

	// handle application credentials update when requested
	// a secret with credentials data is created by the KL apiserver
	// secret deletion marks a request as successful
	credSecrets := &v1.Secret{}
	credSecrets.SetName(kuberlogiccomv1alpha1.CredsUpdateSecretName)
	credSecrets.SetNamespace(kls.Status.Namespace)
	if err := r.Get(ctx, client.ObjectKeyFromObject(credSecrets), credSecrets); err != nil && !k8serrors.IsNotFound(err) {
		log.Error(err, "error getting credentials update secret")
		return ctrl.Result{}, err
	} else if err == nil {
		credMethodRequest := commons.PluginRequestCredentialsMethod{
			Name: kls.GetName(),
			Data: make(map[string]string, len(credSecrets.Data)),
		}

		for k, v := range credSecrets.Data {
			credMethodRequest.Data[k] = string(v)
		}

		m := plugin.GetCredentialsMethod(credMethodRequest)
		if m.Err != "" {
			return ctrl.Result{}, errors.Wrapf(errors.New(m.Err), "failed to get set credentials method")
		}
		if m.Method == "exec" {
			restClient, err := apiutil.RESTClientForGVK(schema.GroupVersionKind{
				Version: "v1",
				Group:   "",
				Kind:    "",
			}, false, r.RESTConfig, serializer.NewCodecFactory(r.Scheme))
			if err != nil {
				return ctrl.Result{}, errors.Wrap(err, "failed to build exec client")
			}

			pods := &v1.PodList{}
			if err := r.List(ctx, pods, client.InNamespace(kls.Status.Namespace), client.MatchingLabels(m.Exec.PodSelector.MatchLabels)); err != nil {
				return ctrl.Result{}, errors.Wrap(err, "failed to list pods")
			}
			if len(pods.Items) != 1 {
				return ctrl.Result{}, errors.Wrapf(err, "exactly one pod should be returned, got %d instead", len(pods.Items))
			}

			pod := pods.Items[0]
			execReq := restClient.Post().Resource("pods").Namespace(pod.GetNamespace()).Name(pod.GetName()).SubResource("exec")
			execReq.VersionedParams(&v1.PodExecOptions{
				Stdin:     false,
				Stdout:    true,
				Stderr:    true,
				TTY:       true,
				Container: m.Exec.Container,
				Command:   m.Exec.Command,
			}, scheme.ParameterCodec)

			stdoutBuf, stderrByf := &bytes.Buffer{}, &bytes.Buffer{}
			exec, err := NewRemoteExecutor(r.RESTConfig, "POST", execReq.URL())
			if err != nil {
				return ctrl.Result{}, errors.Wrapf(err, "failed to create exec executor")
			}
			if err := exec.Stream(remotecommand.StreamOptions{
				Stdin:             nil,
				Stdout:            stdoutBuf,
				Stderr:            stderrByf,
				Tty:               true,
				TerminalSizeQueue: nil,
			}); err != nil {
				stdout, _ := io.ReadAll(stdoutBuf)
				stderr, _ := io.ReadAll(stderrByf)
				log.Error(err, "failed to update user credentials", "stdout", string(stdout), "stderr", string(stderr), "pod", pod.GetName(), "container", m.Exec.Container)
				return ctrl.Result{}, errors.Wrapf(err, "failed to execute update credentials command")
			}
		} else {
			e := errors.New("unknown credentials management method")
			log.Error(e, "", "method", m.Method)
			return ctrl.Result{}, e
		}
		if err := r.Delete(ctx, credSecrets); err != nil {
			log.Error(err, "failed to delete credentials secret request")
			return ctrl.Result{}, err
		}
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
