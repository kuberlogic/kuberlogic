/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package fake

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/testing"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/api"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

// FakeServices implements ServiceInterface
type FakeServices struct {
	Fake *testing.Fake
}

var (
	serviceResource = schema.GroupVersionResource{Group: "kuberlogic.com", Version: "v1alpha1", Resource: "kuberlogicservices"}
	serviceKind     = schema.GroupVersionKind{Group: "kuberlogic.com", Version: "v1alpha1", Kind: "KuberLogicService"}

	_ api.ServiceInterface = &FakeServices{}
)

// Get takes name of the service, and returns the corresponding service object, and an error if there is any.
func (c *FakeServices) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.KuberLogicService, err error) {
	obj, err := c.Fake.Invokes(testing.NewRootGetAction(serviceResource, name), &v1alpha1.KuberLogicService{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KuberLogicService), err
}

// List takes label and field selectors, and returns the list of Services that match those selectors.
func (c *FakeServices) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.KuberLogicServiceList, err error) {
	obj, err := c.Fake.Invokes(testing.NewRootListAction(serviceResource, serviceKind, opts), &v1alpha1.KuberLogicServiceList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.KuberLogicServiceList{ListMeta: obj.(*v1alpha1.KuberLogicServiceList).ListMeta}
	for _, item := range obj.(*v1alpha1.KuberLogicServiceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

//// Watch returns a watch.Interface that watches the requested services.
//func (c *FakeServices) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
//	return c.Fake.
//		InvokesWatch(testing.NewRootWatchAction(serviceResource, opts))
//
//}

// Create takes the representation of a service and creates it.  Returns the server's representation of the service, and an error, if there is any.
func (c *FakeServices) Create(ctx context.Context, pod *v1alpha1.KuberLogicService, opts v1.CreateOptions) (result *v1alpha1.KuberLogicService, err error) {
	obj, err := c.Fake.Invokes(testing.NewRootCreateAction(serviceResource, pod), &v1alpha1.KuberLogicService{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KuberLogicService), err
}

// Delete takes name of the service and deletes it. Returns an error if one occurs.
func (c *FakeServices) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.Invokes(testing.NewRootDeleteAction(serviceResource, name), &v1alpha1.KuberLogicService{})
	return err
}

// Patch applies the patch and returns the patched service.
func (c *FakeServices) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.KuberLogicService, err error) {
	obj, err := c.Fake.Invokes(testing.NewRootPatchSubresourceAction(serviceResource, name, pt, data, subresources...), &v1alpha1.KuberLogicService{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KuberLogicService), err
}
