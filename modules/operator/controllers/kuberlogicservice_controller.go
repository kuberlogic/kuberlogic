package controllers

import (
	"context"
	"github.com/go-logr/logr"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	"github.com/kuberlogic/kuberlogic/modules/operator/monitoring"
	serviceOperator "github.com/kuberlogic/kuberlogic/modules/operator/service-operator"
	"github.com/kuberlogic/kuberlogic/modules/operator/service-operator/interfaces"
	"github.com/kuberlogic/kuberlogic/modules/operator/util"
	mysqlv1 "github.com/presslabs/mysql-operator/pkg/apis/mysql/v1alpha1"
	postgresv1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sync"
	"time"
)

const (
	klsServiceNotReadyDelaySec = 300

	klsFinalizer = kuberlogicv1.Group + "/service-finalizer"
)

// KuberLogicServiceReconciler reconciles a KuberLogicServices object
type KuberLogicServiceReconciler struct {
	client.Client
	Log                 logr.Logger
	Scheme              *runtime.Scheme
	mu                  sync.Mutex
	MonitoringCollector *monitoring.KuberLogicCollector
}

// +kubebuilder:rbac:groups=cloudlinux.com,resources=kuberlogicservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudlinux.com,resources=kuberlogicservices/status,verbs=get;update;patch
func (r *KuberLogicServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("kuberlogicservices", req.NamespacedName)

	defer util.HandlePanic(log)

	mu := getMutex(req.NamespacedName)
	mu.Lock()
	defer mu.Unlock()

	// Fetch the KuberLogicServices instance
	kls := &kuberlogicv1.KuberLogicService{}
	err := r.Get(ctx, req.NamespacedName, kls)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info(req.Namespace, req.Name, " has been deleted")

			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get KuberLogicService")
		return ctrl.Result{}, err
	}
	defer r.MonitoringCollector.MonitorKuberlogicService(kls)

	// fetch tenant information
	kt := new(kuberlogicv1.KuberLogicTenant)
	if err := r.Get(ctx, types.NamespacedName{Name: kls.Namespace, Namespace: ""}, kt); err != nil {
		log.Error(err, "Failed to get kuberlogictenant")
		return ctrl.Result{}, err
	}

	if kls.DeletionTimestamp != nil {
		log.Info("kuberlogicservice is pending for deletion")
		if controllerutil.ContainsFinalizer(kls, klsFinalizer) {
			if err := r.finalize(ctx, kt, kls, log); err != nil {
				log.Error(err, "error finalizing kuberlogicservice")
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(kls, klsFinalizer)
			if err := r.Update(ctx, kls); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	}

	if !controllerutil.ContainsFinalizer(kls, klsFinalizer) {
		log.Info("adding finalizer", "finalizer", klsFinalizer)
		controllerutil.AddFinalizer(kls, klsFinalizer)
		err := r.Update(ctx, kls)
		if err != nil {
			log.Error(err, "error adding finalizer")
		}
		return ctrl.Result{}, err
	}

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
	return r.update(ctx, kt, kls, op, log)
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

	log.Info("ensure that we have dependencies set up")
	if err := r.ensureClusterDependencies(op, kls, ctx); err != nil {
		log.Error(err, "failed to ensure dependencies", "BaseOperator", kls.Spec.Type)
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

func (r *KuberLogicServiceReconciler) update(ctx context.Context, kt *kuberlogicv1.KuberLogicTenant, kls *kuberlogicv1.KuberLogicService, op interfaces.OperatorInterface, log logr.Logger) (reconcile.Result, error) {
	log.Info("Save service to a kuberlogictenant")

	kt.SaveTenantServiceInfo(kls)
	if err := r.Status().Update(ctx, kt); err != nil {
		log.Error(err, "Error updating kuberlogictenant status")
		return ctrl.Result{}, err
	}

	// sync service operator status to kls and check if reconciliation is allowed in this state
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

func (r *KuberLogicServiceReconciler) finalize(ctx context.Context, kt *kuberlogicv1.KuberLogicTenant, kls *kuberlogicv1.KuberLogicService, log logr.Logger) error {
	log.Info("Finalizing service")
	kt.ForgetTenantServiceInfo(kls)
	if err := r.Status().Update(ctx, kt); err != nil {
		return err
	}
	r.MonitoringCollector.ForgetKuberlogicService(kls)
	return nil
}

func syncStatus(kls *kuberlogicv1.KuberLogicService, op interfaces.OperatorInterface) {
	if ready, desc := op.IsReady(); ready {
		kls.MarkReady(desc)
	} else {
		kls.MarkNotReady(desc)
	}
}
