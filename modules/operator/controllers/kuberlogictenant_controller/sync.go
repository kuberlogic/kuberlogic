package kuberlogictenant_controller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// syncer is responsible for controlling all sync operations for a parent kt
type syncer struct {
	kt *kuberlogicv1.KuberLogicTenant

	synced  map[int]client.Object
	syncErr error
	client  client.Client
	scheme  *runtime.Scheme
	log     logr.Logger
	ctx     context.Context
}

const (
	saKey = iota
	nsKey
	imgPullSecretKey
	roleKey
	roleBindingKey
)

func (s *syncer) withNamespace() *syncer {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.kt.GetTenantName(),
		},
	}
	s.syncErr = s.sync(ns, nsKey)
	return s
}

func (s *syncer) withImagePullSecret(parentName, parentNmespace string) *syncer {
	parentSecret := &corev1.Secret{}
	err := s.client.Get(s.ctx, types.NamespacedName{Name: parentName, Namespace: parentNmespace}, parentSecret)
	if err != nil {
		s.syncErr = err
		return s
	}

	clientSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      parentName,
			Namespace: s.kt.GetTenantName(),
		},
		Type: parentSecret.Type,
		Data: parentSecret.Data,
	}
	s.syncErr = s.sync(clientSecret, imgPullSecretKey)
	return s
}

func (s *syncer) withServiceAccount() *syncer {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.kt.GetTenantName(),
			Namespace: s.kt.GetTenantName(),
		},
	}
	s.syncErr = s.sync(sa, saKey)
	return s
}

func (s *syncer) withRole() *syncer {
	r := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.kt.GetTenantName(),
			Namespace: s.kt.GetTenantName(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"get", "list"},
				APIGroups: []string{""},
				Resources: []string{"pods"},
			},
		},
	}
	s.syncErr = s.sync(r, roleKey)
	return s
}

func (s *syncer) withRoleBinding() *syncer {
	role := s.getSyncedObj(roleKey)
	sa := s.getSyncedObj(saKey)

	if role == nil || sa == nil {
		s.syncErr = fmt.Errorf("role or service account must not be nil for rolebinding")
		return s
	}

	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.kt.GetTenantName(),
			Namespace: s.kt.GetTenantName(),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				APIGroup:  "",
				Name:      sa.GetName(),
				Namespace: sa.GetNamespace(),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     role.GetName(),
		},
	}
	s.syncErr = s.sync(rb, roleBindingKey)
	return s
}

// sync function creates or updates v1.Object in cluster
func (s *syncer) sync(object client.Object, key int) error {
	if object == nil {
		s.syncErr = fmt.Errorf("object can't be nil")
	}
	if s.syncErr != nil {
		s.log.Error(s.syncErr, "error happened, sync stopped")
		return s.syncErr
	}

	// guess the type of the object using reflection
	// client.Object is expected to be a pointer here
	res := reflect.TypeOf(object).Elem().Name()
	log := s.log.WithValues("resource", res, "name", object.GetName(), "namespace", object.GetNamespace())
	err := ctrl.SetControllerReference(s.kt, object, s.scheme)
	if err != nil {
		return err
	}

	log.Info("creating object")
	err = s.client.Create(s.ctx, object)
	if k8serrors.IsAlreadyExists(err) {
		log.Info("object already exists. updating")
		err = s.client.Patch(s.ctx, object, client.MergeFrom(object.DeepCopyObject()))
	}
	if err != nil {
		log.Error(err, "error syncing object")
		return err
	}

	s.addSyncedObj(key, object)
	return nil
}

func (s syncer) getSyncedObj(key int) client.Object {
	return s.synced[key]
}

func (s *syncer) addSyncedObj(key int, obj client.Object) {
	s.synced[key] = obj
}

// newSyncer function sync object that will contain all sync information
func newSyncer(ctx context.Context, log logr.Logger, c client.Client, s *runtime.Scheme, kt *kuberlogicv1.KuberLogicTenant, err error) *syncer {
	return &syncer{
		kt:      kt,
		synced:  make(map[int]client.Object),
		syncErr: err,
		client:  c,
		scheme:  s,
		log:     log,
		ctx:     ctx,
	}
}