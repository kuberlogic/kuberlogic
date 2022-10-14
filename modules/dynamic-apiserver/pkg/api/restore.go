/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package api

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

// RestoreGetter has a method to return a RestoreInterface.
// A group's client should implement this interface.
type RestoreGetter interface {
	Restores() RestoreInterface
}

// RestoreInterface has methods to work with Kuberlogic Restores resources.
type RestoreInterface interface {
	Create(ctx context.Context, obj *v1alpha1.KuberlogicServiceRestore, opts v1.CreateOptions) (*v1alpha1.KuberlogicServiceRestore, error)
	Update(ctx context.Context, obj *v1alpha1.KuberlogicServiceRestore, opts v1.UpdateOptions) (*v1alpha1.KuberlogicServiceRestore, error)
	//UpdateStatus(ctx context.Context, Restores *v1alpha1.KuberlogicServiceRestore, opts v1.UpdateOptions) (*v1alpha1.KuberlogicServiceRestore, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	//DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.KuberlogicServiceRestore, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.KuberlogicServiceRestoreList, error)
	//Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.KuberlogicServiceRestore, err error)
}

const restoreK8sResource = "kuberlogicservicerestores"

type Restores struct {
	restClient rest.Interface
}

var _ RestoreInterface = &Restores{}

// NewRestores returns a Restores
func NewRestores(c rest.Interface) RestoreInterface {
	return &Restores{
		restClient: c,
	}
}

func (svc *Restores) Create(ctx context.Context, obj *v1alpha1.KuberlogicServiceRestore, opts v1.CreateOptions) (*v1alpha1.KuberlogicServiceRestore, error) {
	result := &v1alpha1.KuberlogicServiceRestore{}
	err := svc.restClient.Post().
		Resource(restoreK8sResource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	return result, err
}

func (svc *Restores) Update(ctx context.Context, obj *v1alpha1.KuberlogicServiceRestore, opts v1.UpdateOptions) (*v1alpha1.KuberlogicServiceRestore, error) {
	result := &v1alpha1.KuberlogicServiceRestore{}
	err := svc.restClient.Put().
		Resource(restoreK8sResource).
		Name(obj.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	return result, err
}

func (svc *Restores) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (*v1alpha1.KuberlogicServiceRestore, error) {
	result := &v1alpha1.KuberlogicServiceRestore{}
	err := svc.restClient.Patch(pt).
		Resource(restoreK8sResource).
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return result, err
}

func (svc *Restores) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return svc.restClient.Delete().
		Resource(restoreK8sResource).
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

func (svc *Restores) Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.KuberlogicServiceRestore, error) {
	result := &v1alpha1.KuberlogicServiceRestore{}
	err := svc.restClient.Get().
		Resource(restoreK8sResource).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return result, err
}
func (svc *Restores) List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.KuberlogicServiceRestoreList, error) {
	result := &v1alpha1.KuberlogicServiceRestoreList{}
	err := svc.restClient.Get().
		Resource(restoreK8sResource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return result, err
}
