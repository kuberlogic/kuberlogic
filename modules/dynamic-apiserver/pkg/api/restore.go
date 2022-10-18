/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package api

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

// RestoreGetter has a method to return a RestoreInterface.
// A group's client should implement this interface.
type RestoreGetter interface {
	Restores() RestoreInterface
}

// RestoreInterface has methods to work with Kuberlogic restores resources.
type RestoreInterface interface {
	Create(ctx context.Context, obj *v1alpha1.KuberlogicServiceRestore, opts v1.CreateOptions) (*v1alpha1.KuberlogicServiceRestore, error)
	Update(ctx context.Context, obj *v1alpha1.KuberlogicServiceRestore, opts v1.UpdateOptions) (*v1alpha1.KuberlogicServiceRestore, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.KuberlogicServiceRestore, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.KuberlogicServiceRestoreList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.KuberlogicServiceRestore, err error)
}

const restoreK8sResource = "kuberlogicservicerestores"

type restores struct {
	restClient rest.Interface
}

var _ RestoreInterface = &restores{}

// NewRestores returns a restores
func NewRestores(c rest.Interface) RestoreInterface {
	return &restores{
		restClient: c,
	}
}

func (r *restores) Create(ctx context.Context, obj *v1alpha1.KuberlogicServiceRestore, opts v1.CreateOptions) (*v1alpha1.KuberlogicServiceRestore, error) {
	result := &v1alpha1.KuberlogicServiceRestore{}
	err := r.restClient.Post().
		Resource(restoreK8sResource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	return result, err
}

func (r *restores) Update(ctx context.Context, obj *v1alpha1.KuberlogicServiceRestore, opts v1.UpdateOptions) (*v1alpha1.KuberlogicServiceRestore, error) {
	result := &v1alpha1.KuberlogicServiceRestore{}
	err := r.restClient.Put().
		Resource(restoreK8sResource).
		Name(obj.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	return result, err
}

func (r *restores) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (*v1alpha1.KuberlogicServiceRestore, error) {
	result := &v1alpha1.KuberlogicServiceRestore{}
	err := r.restClient.Patch(pt).
		Resource(restoreK8sResource).
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return result, err
}

func (r *restores) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return r.restClient.Delete().
		Resource(restoreK8sResource).
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

func (r *restores) Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.KuberlogicServiceRestore, error) {
	result := &v1alpha1.KuberlogicServiceRestore{}
	err := r.restClient.Get().
		Resource(restoreK8sResource).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return result, err
}
func (r *restores) List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.KuberlogicServiceRestoreList, error) {
	result := &v1alpha1.KuberlogicServiceRestoreList{}
	err := r.restClient.Get().
		Resource(restoreK8sResource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return result, err
}

// Watch returns a watch.Interface that watches the requested backups.
func (r *restores) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return r.restClient.Get().
		Resource(restoreK8sResource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}
