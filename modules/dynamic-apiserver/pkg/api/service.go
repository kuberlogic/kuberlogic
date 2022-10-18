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

// ServiceGetter has a method to return a ServiceInterface.
// A group's client should implement this interface.
type ServiceGetter interface {
	Services() ServiceInterface
}

// ServiceInterface has methods to work with Kuberlogic services resources.
type ServiceInterface interface {
	Create(ctx context.Context, service *v1alpha1.KuberLogicService, opts v1.CreateOptions) (*v1alpha1.KuberLogicService, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.KuberLogicService, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.KuberLogicServiceList, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.KuberLogicService, err error)
}

const serviceK8sResource = "kuberlogicservices"

type services struct {
	restClient rest.Interface
}

var _ ServiceInterface = &services{}

// NewServices returns a services
func NewServices(c rest.Interface) ServiceInterface {
	return &services{
		restClient: c,
	}
}

func (svc *services) Create(ctx context.Context, service *v1alpha1.KuberLogicService, opts v1.CreateOptions) (*v1alpha1.KuberLogicService, error) {
	result := &v1alpha1.KuberLogicService{}
	err := svc.restClient.Post().
		Resource(serviceK8sResource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(service).
		Do(ctx).
		Into(result)
	return result, err
}

func (svc *services) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (*v1alpha1.KuberLogicService, error) {
	result := &v1alpha1.KuberLogicService{}
	err := svc.restClient.Patch(pt).
		Resource(serviceK8sResource).
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return result, err
}

func (svc *services) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return svc.restClient.Delete().
		Resource(serviceK8sResource).
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

func (svc *services) Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.KuberLogicService, error) {
	result := &v1alpha1.KuberLogicService{}
	err := svc.restClient.Get().
		Resource(serviceK8sResource).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return result, err
}
func (svc *services) List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.KuberLogicServiceList, error) {
	result := &v1alpha1.KuberLogicServiceList{}
	err := svc.restClient.Get().
		Resource(serviceK8sResource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return result, err
}
