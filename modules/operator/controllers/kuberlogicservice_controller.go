package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/monitoring"
	"github.com/kuberlogic/operator/modules/operator/service-operator"
	mysqlv1 "github.com/presslabs/mysql-operator/pkg/apis/mysql/v1alpha1"
	redisv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	postgresv1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

// KuberLogicServiceReconciler reconciles a KuberLogicServices object
type KuberLogicServiceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	mu     sync.Mutex
}

// +kubebuilder:rbac:groups=cloudlinux.com,resources=kuberlogicservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudlinux.com,resources=kuberlogicservices/status,verbs=get;update;patch
func (r *KuberLogicServiceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("kuberlogicservices", req.NamespacedName)

	r.mu.Lock()
	defer r.mu.Unlock()

	// metrics key
	monitoringKey := fmt.Sprintf("%s/%s", req.Name, req.Namespace)

	// Fetch the KuberLogicServices instance
	kls := &kuberlogicv1.KuberLogicService{}
	err := r.Get(ctx, req.NamespacedName, kls)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info(req.Namespace, req.Name, " has been deleted")
			delete(monitoring.KuberLogicServices, monitoringKey)
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get KuberLogicService")
		return ctrl.Result{}, err
	}

	op, err := service_operator.GetOperator(kls.Spec.Type)
	if err != nil {
		log.Error(err, "Could not define the base operator")
		return ctrl.Result{}, err
	}

	if kls.InitDefaults(op.GetDefaults()) {
		err = r.Update(ctx, kls)
		if err != nil {
			log.Error(err, "Failed to update KuberLogicService")
			return ctrl.Result{}, err
		} else {
			log.Info("KuberLogicService defaults is updated")
		}
	}

	found := op.AsRuntimeObject()
	err = r.Get(
		ctx,
		types.NamespacedName{
			Name:      op.Name(kls),
			Namespace: kls.Namespace,
		},
		found,
	)

	if err != nil && k8serrors.IsNotFound(err) {
		// Define a new cluster
		dep, err := r.defineCluster(op, kls)
		if err != nil {
			log.Error(err, "Could not generate definition struct", "Operator", kls.Spec.Type)
			return ctrl.Result{}, err
		}

		log.Info("Creating a new Cluster", "Operator", kls.Spec.Type)
		err = r.Create(ctx, dep)
		if err != nil && k8serrors.IsAlreadyExists(err) {
			log.Info("Cluster already exists", "Operator", kls.Spec.Type)
		} else if err != nil {
			log.Error(err, "Failed to create new Cluster", "Operator", kls.Spec.Type)
			return ctrl.Result{}, err
		} else {
			// cluster created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		}
	} else if err != nil {
		log.Error(err, "Failed to get object", "Operator", kls.Spec.Type)
		return ctrl.Result{}, err
	}

	op.InitFrom(found)

	log.Info("ensure that we have dependencies set up")
	if err := r.ensureClusterDependencies(op, kls, ctx); err != nil {
		log.Error(err, "failed to ensure dependencies", "Operator", kls.Spec.Type)
		return ctrl.Result{}, err
	}

	if !op.IsEqual(kls) {
		op.Update(kls)

		err = r.Update(ctx, op.AsRuntimeObject())
		if err != nil {
			log.Error(err, "Failed to update object", "Operator", kls.Spec.Type)
			return ctrl.Result{}, err
		} else {
			log.Info("Cluster is updated", "Operator", kls.Spec.Type)
			return ctrl.Result{}, nil
		}
	}
	log.Info("No difference", "Operator", kls.Spec.Type)

	status := op.CurrentStatus()
	if !kls.IsEqual(status) {
		kls.SetStatus(status)
		err = r.Update(ctx, kls)
		//err = r.Status().Update(ctx, kls) # FIXME: Figure out why it's failed
		if err != nil {
			log.Error(err, "Failed to update kls object")
			return ctrl.Result{}, err
		} else {
			log.Info("KuberLogicService status is updated", "Status", kls.GetStatus())
		}
	}

	monitoring.KuberLogicServices[monitoringKey] = kls

	return ctrl.Result{}, nil
}

func (r *KuberLogicServiceReconciler) ensureClusterDependencies(op service_operator.Operator, cm *kuberlogicv1.KuberLogicService, ctx context.Context) error {
	credSecret, err := op.GetCredentialsSecret()
	if err != nil {
		return err
	}
	if credSecret != nil {
		if err := ctrl.SetControllerReference(cm, credSecret, r.Scheme); err != nil {
			return err
		}
		if err := r.Create(ctx, credSecret); err != nil && !k8serrors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

func (r *KuberLogicServiceReconciler) defineCluster(op service_operator.Operator, cm *kuberlogicv1.KuberLogicService) (runtime.Object, error) {
	op.Init(cm)
	op.Update(cm)

	// Set KuberLogicService instance as the owner and controller
	// if KuberLogicService will remove -> dep also should be removed automatically
	err := ctrl.SetControllerReference(cm, op.AsMetaObject(), r.Scheme)
	if err != nil {
		return nil, err
	}

	return op.AsRuntimeObject(), nil
}

func (r *KuberLogicServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kuberlogicv1.KuberLogicService{}).
		Owns(&mysqlv1.MysqlCluster{}).
		Owns(&redisv1.RedisFailover{}).
		Owns(&postgresv1.Postgresql{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
