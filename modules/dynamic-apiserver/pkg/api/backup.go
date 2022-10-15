/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package api

import (
	"context"
	"k8s.io/apimachinery/pkg/watch"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

// BackupGetter has a method to return a BackupInterface.
// A group's client should implement this interface.
type BackupGetter interface {
	Backups() BackupInterface
}

// BackupInterface has methods to work with Kuberlogic backups resources.
type BackupInterface interface {
	Create(ctx context.Context, backup *v1alpha1.KuberlogicServiceBackup, opts v1.CreateOptions) (*v1alpha1.KuberlogicServiceBackup, error)
	Update(ctx context.Context, backup *v1alpha1.KuberlogicServiceBackup, opts v1.UpdateOptions) (*v1alpha1.KuberlogicServiceBackup, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.KuberlogicServiceBackup, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.KuberlogicServiceBackupList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.KuberlogicServiceBackup, err error)
}

const backupK8sResource = "kuberlogicservicebackups"

type backups struct {
	restClient rest.Interface
}

var _ BackupInterface = &backups{}

// NewBackups returns a backups
func NewBackups(c rest.Interface) BackupInterface {
	return &backups{
		restClient: c,
	}
}

func (b *backups) Create(ctx context.Context, backup *v1alpha1.KuberlogicServiceBackup, opts v1.CreateOptions) (*v1alpha1.KuberlogicServiceBackup, error) {
	result := &v1alpha1.KuberlogicServiceBackup{}
	err := b.restClient.Post().
		Resource(backupK8sResource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(backup).
		Do(ctx).
		Into(result)
	return result, err
}

func (b *backups) Update(ctx context.Context, backup *v1alpha1.KuberlogicServiceBackup, opts v1.UpdateOptions) (*v1alpha1.KuberlogicServiceBackup, error) {
	result := &v1alpha1.KuberlogicServiceBackup{}
	err := b.restClient.Put().
		Resource(backupK8sResource).
		Name(backup.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(backup).
		Do(ctx).
		Into(result)
	return result, err
}

func (b *backups) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (*v1alpha1.KuberlogicServiceBackup, error) {
	result := &v1alpha1.KuberlogicServiceBackup{}
	err := b.restClient.Patch(pt).
		Resource(backupK8sResource).
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return result, err
}

func (b *backups) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return b.restClient.Delete().
		Resource(backupK8sResource).
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

func (b *backups) Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.KuberlogicServiceBackup, error) {
	result := &v1alpha1.KuberlogicServiceBackup{}
	err := b.restClient.Get().
		Resource(backupK8sResource).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return result, err
}
func (b *backups) List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.KuberlogicServiceBackupList, error) {
	result := &v1alpha1.KuberlogicServiceBackupList{}
	err := b.restClient.Get().
		Resource(backupK8sResource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return result, err
}

// Watch returns a watch.Interface that watches the requested backups.
func (b *backups) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return b.restClient.Get().
		Resource(backupK8sResource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}
