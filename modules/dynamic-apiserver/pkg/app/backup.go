/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package app

import (
	"context"
	"fmt"
	"sort"
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

type ExtendedBackupGetter interface {
	Backups() ExtendedBackupInterface
}

type ExtendedBackupInterface interface {
	api.BackupInterface
	CreateByServiceName(ctx context.Context, name string) (*v1alpha1.KuberlogicServiceBackup, error)
	FirstSuccessful(ctx context.Context, opt v1.ListOptions, sortBy func([]*v1alpha1.KuberlogicServiceBackup) sort.Interface) (*v1alpha1.KuberlogicServiceBackup, error)
	Wait(ctx context.Context, resource *v1alpha1.KuberlogicServiceBackup, condition func(event watch.Event) (bool, error), timeout time.Duration) (*v1alpha1.KuberlogicServiceBackup, error)
}
type backups struct {
	api.BackupInterface
}

var _ ExtendedBackupInterface = &backups{}

func newBackups(c rest.Interface) ExtendedBackupInterface {
	s := &backups{}
	s.BackupInterface = api.NewBackups(c)
	return s
}

func (b *backups) CreateByServiceName(ctx context.Context, name string) (*v1alpha1.KuberlogicServiceBackup, error) {
	klb := &v1alpha1.KuberlogicServiceBackup{
		ObjectMeta: v1.ObjectMeta{
			Name: fmt.Sprintf("%s-%d", name, time.Now().Unix()),
			Labels: map[string]string{
				util.BackupRestoreServiceField: name,
			},
		},
		Spec: v1alpha1.KuberlogicServiceBackupSpec{
			KuberlogicServiceName: name,
		},
	}
	return b.Create(ctx, klb, v1.CreateOptions{})
}

func (b *backups) Wait(
	ctx context.Context,
	resource *v1alpha1.KuberlogicServiceBackup,
	condition func(event watch.Event) (bool, error),
	timeout time.Duration,
) (*v1alpha1.KuberlogicServiceBackup, error) {
	ctx, cancel := watchtools.ContextWithOptionalTimeout(ctx, timeout)
	defer cancel()

	event, err := watchtools.UntilWithSync(ctx, &cache.ListWatch{
		ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
			return b.List(ctx, options)
		},
		WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
			return b.Watch(ctx, options)
		},
	}, resource, nil, condition)
	if err != nil {
		return nil, errors.Wrap(err, "error waiting condition")
	}
	obj, ok := event.Object.(*v1alpha1.KuberlogicServiceBackup)
	if !ok {
		return nil, errors.Wrap(err, "error conversion for backup object")
	}
	return obj, nil
}

func (b *backups) FirstSuccessful(ctx context.Context, opt v1.ListOptions, sortBy func([]*v1alpha1.KuberlogicServiceBackup) sort.Interface) (*v1alpha1.KuberlogicServiceBackup, error) {
	r, err := b.List(ctx, opt)
	if err != nil {
		return nil, errors.Wrap(err, "error listing service backups")
	}
	listOfBackups := make([]*v1alpha1.KuberlogicServiceBackup, 0)
	for _, backup := range r.Items {
		listOfBackups = append(listOfBackups, &backup)
	}
	sort.Sort(sortBy(listOfBackups))

	for _, backup := range listOfBackups {
		if backup.IsSuccessful() {
			return backup, nil
		}
	}
	return nil, errors.New("No successful backup found")
}
