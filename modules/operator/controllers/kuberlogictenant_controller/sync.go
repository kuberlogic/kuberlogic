package kuberlogictenant_controller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/rbac/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type syncer struct {
	kt *kuberlogicv1.KuberLogicTenant

	synced map[int]client.Object
	syncErr error
	client client.Client
	scheme *runtime.Scheme
	log logr.Logger
	ctx context.Context
}

const (
	saKey = iota
	nsKey
	imgPullSecretKey
	roleKey
	roleBindingKey
)

func (s *syncer) withNamespace() *syncer {
	s.log.Info("syncing tenant namespace")
	ns := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: s.kt.GetTenantName(),
		},
	}
	s.syncErr = s.sync(ns, nsKey)
	return s
}

func (s *syncer) withImagePullSecret(parentName, parentNmespace string) *syncer {
	s.log.Info("syncing tenant image pull secret")
	parentSecret := &corev1.Secret{}
	err := s.client.Get(s.ctx, types.NamespacedName{Name: parentName, Namespace: parentNmespace}, parentSecret)
	if err != nil {
		s.syncErr = err
		return s
	}

	clientSecret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name: parentName,
			Namespace: s.kt.GetTenantName(),
		},
		Type: parentSecret.Type,
		Data: parentSecret.Data,
	}
	s.syncErr = s.sync(clientSecret, imgPullSecretKey)
	return s
}

func (s *syncer) withServiceAccount() *syncer {
	s.log.Info("syncing tenant service account")
	sa := &corev1.ServiceAccount{
		ObjectMeta: v1.ObjectMeta{
			Name: s.kt.GetTenantName(),
			Namespace: s.kt.GetTenantName(),
		},
	}
	s.syncErr = s.sync(sa, saKey)
	return s
}

func (s *syncer) withRole() *syncer {
	s.log.Info("syncing tenant service account role")
	r := &v1beta1.Role{
		ObjectMeta: v1.ObjectMeta{
			Name: s.kt.GetTenantName(),
			Namespace: s.kt.GetTenantName(),
		},
		Rules: []v1beta1.PolicyRule{
			v1beta1.PolicyRule{
				Verbs:           []string{"get", "list"},
				APIGroups:       []string{""},
				Resources:       []string{"pods"},
			},
		},
	}
	s.syncErr = s.sync(r, roleKey)
	return s
}

func (s *syncer) withRoleBinding() *syncer {
	s.log.Info("syncing tenant service account role binding")

	role := s.getSyncedObj(roleKey)
	sa := s.getSyncedObj(saKey)

	if role == nil || sa == nil {
		s.syncErr = fmt.Errorf("role or service account must not be nil for rolebinding")
		return s
	}

	rb := &v1beta1.RoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Name: s.kt.GetTenantName(),
			Namespace: s.kt.GetTenantName(),
		},
		Subjects:   []v1beta1.Subject{
			{
				Kind:      "ServiceAccount",
				APIGroup:  "v1",
				Name:      sa.GetName(),
				Namespace: sa.GetNamespace(),
						},
		},
		RoleRef:    v1beta1.RoleRef{
			APIGroup: "v1beta1",
			Kind:     "Role",
			Name:     role.GetName(),
		},
	}
	s.syncErr = s.sync(rb, roleBindingKey)
	return s
}

// sync function creates or updates client.Object in cluster
func (s *syncer) sync(object client.Object, key int) error {
	if object == nil {
		s.syncErr = fmt.Errorf("object can't be nil")
	}
	if s.syncErr != nil {
		s.log.Error(s.syncErr, "error happened, sync stopped")
		return s.syncErr
	}

	err := ctrl.SetControllerReference(s.kt, object, s.scheme)
	if err != nil {
		return err
	}

	err = s.client.Create(s.ctx, object)
	if k8serrors.IsAlreadyExists(err) {
		err = s.client.Patch(s.ctx, object, client.MergeFrom(object.DeepCopyObject()))
	}
	if err != nil {
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

func newSyncer(ctx context.Context, log logr.Logger, c client.Client, s *runtime.Scheme, kt *kuberlogicv1.KuberLogicTenant, err error) *syncer {
	return &syncer{
		kt:      kt,
		syncErr: err,
		client:  c,
		scheme:  s,
		log:     log,
		ctx:     ctx,
	}
}
