/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package app

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/api"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

type ExtendedRestoreGetter interface {
	Restores() ExtendedRestoreInterface
}

type ExtendedRestoreInterface interface {
	api.RestoreInterface
	CreateByBackupName(ctx context.Context, name string) (*v1alpha1.KuberlogicServiceRestore, error)
	Wait(ctx context.Context, resource *v1alpha1.KuberlogicServiceRestore, condition func(event watch.Event) (bool, error), timeout time.Duration) (*v1alpha1.KuberlogicServiceRestore, error)
}

type restores struct {
	api.RestoreInterface
}

var _ ExtendedRestoreInterface = &restores{}

func newRestores(c rest.Interface) ExtendedRestoreInterface {
	s := &restores{}
	s.RestoreInterface = api.NewRestores(c)
	return s
}

func (r *restores) CreateByBackupName(ctx context.Context, name string) (*v1alpha1.KuberlogicServiceRestore, error) {
	restore := &v1alpha1.KuberlogicServiceRestore{
		ObjectMeta: v1.ObjectMeta{
			Name: fmt.Sprintf("%s-%d", name, time.Now().Unix()),
			Labels: map[string]string{
				util.BackupRestoreServiceField: name,
			},
		},
		Spec: v1alpha1.KuberlogicServiceRestoreSpec{
			KuberlogicServiceBackup: name,
		},
	}
	return r.Create(ctx, restore, v1.CreateOptions{})
}

func (r *restores) Wait(
	ctx context.Context,
	resource *v1alpha1.KuberlogicServiceRestore,
	condition func(event watch.Event) (bool, error),
	timeout time.Duration,
) (*v1alpha1.KuberlogicServiceRestore, error) {
	ctx, cancel := watchtools.ContextWithOptionalTimeout(ctx, timeout)
	defer cancel()

	event, err := watchtools.UntilWithSync(ctx, &cache.ListWatch{
		ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
			return r.List(ctx, options)
		},
		WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
			return r.Watch(ctx, options)
		},
	}, resource, nil, condition)
	if err != nil {
		return nil, errors.Wrap(err, "error waiting condition")
	}
	obj, ok := event.Object.(*v1alpha1.KuberlogicServiceRestore)
	if !ok {
		return nil, errors.Wrap(err, "error conversion for backup object")
	}
	return obj, nil
}
