package controllers

import (
	"context"
	"github.com/go-logr/logr"
	mysqlv1 "github.com/presslabs/mysql-operator/pkg/apis/mysql/v1alpha1"
	redisv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	postgresv1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	"gitlab.com/cloudmanaged/operator/api/v1/operator"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

// CloudManagedReconciler reconciles a CloudManaged object
type CloudManagedReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	mu     sync.Mutex
}

// +kubebuilder:rbac:groups=cloudlinux.com,resources=cloudmanageds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudlinux.com,resources=cloudmanageds/status,verbs=get;update;patch
func (r *CloudManagedReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("cloudmanaged", req.NamespacedName)

	r.mu.Lock()
	defer r.mu.Unlock()

	// Fetch the Cloudmanaged instance
	cloudmanaged := &cloudlinuxv1.CloudManaged{}
	err := r.Get(ctx, req.NamespacedName, cloudmanaged)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Cloudmanaged")
		return ctrl.Result{}, err
	}

	op, err := operator.GetOperator(cloudmanaged.Spec.Type)
	if err != nil {
		log.Error(err, "Could not define the base operator")
		return ctrl.Result{}, err
	}

	if cloudmanaged.InitDefaults(op.GetDefaults()) {
		err = r.Update(ctx, cloudmanaged)
		if err != nil {
			log.Error(err, "Failed to update cloudmanaged object")
			return ctrl.Result{}, err
		} else {
			log.Info("Cloudmanaged defaults is updated")
		}
	}

	found := op.AsRuntimeObject()
	err = r.Get(
		ctx,
		types.NamespacedName{
			Name:      cloudmanaged.Name,
			Namespace: cloudmanaged.Namespace,
		},
		found,
	)

	if err != nil && k8serrors.IsNotFound(err) {
		// Define a new cluster
		dep, err := r.defineCluster(op, cloudmanaged)
		if err != nil {
			log.Error(err, "Could not generate definition struct", "Operator", cloudmanaged.Spec.Type)
			return ctrl.Result{}, err
		}

		log.Info("Creating a new Cluster", "Operator", cloudmanaged.Spec.Type)
		err = r.Create(ctx, dep)
		if err != nil && k8serrors.IsAlreadyExists(err) {
			log.Info("Cluster already exists", "Operator", cloudmanaged.Spec.Type)
		} else if err != nil {
			log.Error(err, "Failed to create new Cluster", "Operator", cloudmanaged.Spec.Type)
			return ctrl.Result{}, err
		} else {
			// cluster created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		}
	} else if err != nil {
		log.Error(err, "Failed to get object", "Operator", cloudmanaged.Spec.Type)
		return ctrl.Result{}, err
	}

	op.InitFrom(found)
	if !op.IsEqual(cloudmanaged) {
		op.Update(cloudmanaged)

		err = r.Update(ctx, op.AsRuntimeObject())
		if err != nil {
			log.Error(err, "Failed to update object", "Operator", cloudmanaged.Spec.Type)
			return ctrl.Result{}, err
		} else {
			log.Info("Cluster is updated", "Operator", cloudmanaged.Spec.Type)
		}
	} else {
		log.Info("No difference", "Operator", cloudmanaged.Spec.Type)
	}

	status := op.CurrentStatus()
	if !cloudmanaged.IsEqual(status) {
		cloudmanaged.SetStatus(status)
		err = r.Update(ctx, cloudmanaged)
		//err = r.Status().Update(ctx, cloudmanaged) # FIXME: Figure out why it's failed
		if err != nil {
			log.Error(err, "Failed to update cloudmanaged object")
		} else {
			log.Info("Cloudmanaged status is updated", "Status", cloudmanaged.GetStatus())
		}
	}

	cloudmanaged.SetMetrics()

	return ctrl.Result{}, nil
}

func (r *CloudManagedReconciler) defineCluster(op operator.Operator, cm *cloudlinuxv1.CloudManaged) (runtime.Object, error) {
	op.Init(cm)
	op.Update(cm)

	// Set cloudmanage instance as the owner and controller
	// if cloudmanage will remove -> dep also should be removed automatically
	err := ctrl.SetControllerReference(cm, op.AsMetaObject(), r.Scheme)
	if err != nil {
		return nil, err
	}

	return op.AsRuntimeObject(), nil
}

func (r *CloudManagedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudlinuxv1.CloudManaged{}).
		Owns(&mysqlv1.MysqlCluster{}).
		Owns(&redisv1.RedisFailover{}).
		Owns(&postgresv1.Postgresql{}).
		Complete(r)
}
