package controllers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/cfg"
	"github.com/kuberlogic/operator/modules/operator/monitoring"
	serviceOperator "github.com/kuberlogic/operator/modules/operator/service-operator"
	"github.com/kuberlogic/operator/modules/operator/service-operator/interfaces"
	"github.com/kuberlogic/operator/modules/operator/util"
	mysqlv1 "github.com/presslabs/mysql-operator/pkg/apis/mysql/v1alpha1"
	postgresv1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sync"
	"time"
)

const (
	klsServiceNotReadyDelaySec = 300
)

// KuberLogicServiceReconciler reconciles a KuberLogicServices object
type KuberLogicServiceReconciler struct {
	client.Client
	Log                 logr.Logger
	Scheme              *runtime.Scheme
	mu                  sync.Mutex
	MonitoringCollector *monitoring.KuberLogicCollector
	Config              *cfg.Config
}

// +kubebuilder:rbac:groups=cloudlinux.com,resources=kuberlogicservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudlinux.com,resources=kuberlogicservices/status,verbs=get;update;patch
func (r *KuberLogicServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("kuberlogicservices", req.NamespacedName)

	defer util.HandlePanic(log)

	mu := getMutex(req.NamespacedName)
	mu.Lock()
	defer mu.Unlock()

	// metrics key
	monitoringKey := fmt.Sprintf(req.String())

	// Fetch the KuberLogicServices instance
	kls := &kuberlogicv1.KuberLogicService{}
	err := r.Get(ctx, req.NamespacedName, kls)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info(req.Namespace, req.Name, " has been deleted")
			r.MonitoringCollector.ForgetKuberlogicService(monitoringKey)
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get KuberLogicService")
		return ctrl.Result{}, err
	}
	defer r.MonitoringCollector.MonitorKuberlogicService(monitoringKey, kls)

	op, err := serviceOperator.GetOperator(kls.Spec.Type)
	if err != nil {
		log.Error(err, "Could not define the base operator")
		return ctrl.Result{}, err
	}

	serviceObj := op.AsClientObject()
	err = r.Get(
		ctx,
		types.NamespacedName{
			Name:      op.Name(kls),
			Namespace: kls.Namespace,
		},
		serviceObj,
	)

	if err != nil && k8serrors.IsNotFound(err) {
		return r.create(ctx, kls, op, log)
	} else if err != nil {
		log.Error(err, "Failed to get object", "BaseOperator", kls.Spec.Type)
		return ctrl.Result{}, err
	}
	op.InitFrom(serviceObj)

	log.Info("ensure that we have operator dependencies set up")
	if err := r.ensureClusterDependencies(op, kls, ctx); err != nil {
		log.Error(err, "failed to ensure dependencies", "BaseOperator", kls.Spec.Type)
		return ctrl.Result{}, err
	}

	return r.update(ctx, kls, op, log)
}

func (r *KuberLogicServiceReconciler) ensureClusterDependencies(op interfaces.OperatorInterface, kls *kuberlogicv1.KuberLogicService, ctx context.Context) error {
	credSecret, err := op.GetInternalDetails().GetCredentialsSecret()
	if err != nil {
		return err
	}
	if credSecret != nil {
		if err := ctrl.SetControllerReference(kls, credSecret, r.Scheme); err != nil {
			return err
		}
		if err := r.Create(ctx, credSecret); err != nil && !k8serrors.IsAlreadyExists(err) {
			return err
		}
	}

	// sync imagePullSecret
	imgPullSecret := new(corev1.Secret)
	imgPullSecretName := types.NamespacedName{
		Namespace: r.Config.Namespace,
		Name:      r.Config.ImagePullSecretName,
	}
	if err := r.Get(ctx, imgPullSecretName, imgPullSecret); err != nil {
		return err
	}
	nsPullSecret := new(corev1.Secret)
	imgPullSecretName.Namespace = kls.Namespace
	if err := r.Get(ctx, imgPullSecretName, nsPullSecret); err != nil {
		if k8serrors.IsNotFound(err) {
			// this object doesn't have ownerReference set to kls intentionally!
			// as we don't wnt to clean it up if only one kls is deleted
			return r.Create(ctx, &corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name:      imgPullSecretName.Name,
					Namespace: kls.Namespace,
				},
				Data: imgPullSecret.Data,
			})
		}
		return err
	}

	dockerCfgKey := ".dockerconfigjson"
	if cmp := bytes.Compare(imgPullSecret.Data[dockerCfgKey], nsPullSecret.Data[dockerCfgKey]); cmp != 0 {
		nsPullSecret.Data[dockerCfgKey] = imgPullSecret.Data[dockerCfgKey]
		return r.Update(ctx, nsPullSecret)
	}
	return nil
}

func (r *KuberLogicServiceReconciler) defineCluster(op interfaces.OperatorInterface, cm *kuberlogicv1.KuberLogicService) (client.Object, error) {
	op.Init(cm)
	op.Update(cm)

	// Set KuberLogicService instance as the owner and controller
	// if KuberLogicService will remove -> dep also should be removed automatically
	err := ctrl.SetControllerReference(cm, op.AsMetaObject(), r.Scheme)
	if err != nil {
		return nil, err
	}

	return op.AsClientObject(), nil
}

func (r *KuberLogicServiceReconciler) create(ctx context.Context, kls *kuberlogicv1.KuberLogicService, op interfaces.OperatorInterface, log logr.Logger) (reconcile.Result, error) {
	dep, err := r.defineCluster(op, kls)
	if err != nil {
		log.Error(err, "Could not generate definition struct", "BaseOperator", kls.Spec.Type)
		return ctrl.Result{}, err
	}

	log.Info("Creating a new Cluster", "BaseOperator", kls.Spec.Type)
	if err := r.Create(ctx, dep); err != nil && k8serrors.IsAlreadyExists(err) {
		log.Info("Cluster already exists", "BaseOperator", kls.Spec.Type)
	} else if err != nil {
		log.Error(err, "Failed to create new Cluster", "BaseOperator", kls.Spec.Type)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *KuberLogicServiceReconciler) update(ctx context.Context, kls *kuberlogicv1.KuberLogicService, op interfaces.OperatorInterface, log logr.Logger) (reconcile.Result, error) {
	// sync status first
	syncStatus(kls, op)
	if err := r.Status().Update(ctx, kls); err != nil {
		log.Error(err, "error updating status")
		return ctrl.Result{}, err
	}
	if !kls.ReconciliationAllowed() {
		log.Info("updates are not allowed in current service state")
		return ctrl.Result{
			RequeueAfter: time.Second * klsServiceNotReadyDelaySec,
		}, nil
	}

	op.Update(kls)
	if err := r.Update(ctx, op.AsClientObject()); err != nil {
		log.Error(err, "Failed to update object", "BaseOperator", kls.Spec.Type)
		return ctrl.Result{}, err
	}
	log.Info("Cluster is updated", "BaseOperator", kls.Spec.Type)
	return ctrl.Result{}, nil
}

func (r *KuberLogicServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kuberlogicv1.KuberLogicService{}).
		Owns(&mysqlv1.MysqlCluster{}).
		Owns(&postgresv1.Postgresql{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

func syncStatus(kls *kuberlogicv1.KuberLogicService, op interfaces.OperatorInterface) {
	if ready, desc := op.IsReady(); ready {
		kls.MarkReady(desc)
	} else {
		kls.MarkNotReady(desc)
	}
}
