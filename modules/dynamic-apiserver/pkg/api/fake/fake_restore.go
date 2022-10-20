/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package fake

import (
	"context"
	"k8s.io/apimachinery/pkg/watch"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/testing"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/api"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

// FakeRestores implements RestoreInterface
type FakeRestores struct {
	Fake *testing.Fake
}

var (
	restoreResource = schema.GroupVersionResource{Group: "kuberlogic.com", Version: "v1alpha1", Resource: "kuberlogicservicerestores"}
	restoreKind     = schema.GroupVersionKind{Group: "kuberlogic.com", Version: "v1alpha1", Kind: "KuberlogicServiceRestore"}

	_ api.RestoreInterface = &FakeRestores{}
)

// Get takes name of the service, and returns the corresponding restore object, and an error if there is any.
func (c *FakeRestores) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.KuberlogicServiceRestore, err error) {
	obj, err := c.Fake.Invokes(testing.NewRootGetAction(restoreResource, name), &v1alpha1.KuberlogicServiceRestore{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KuberlogicServiceRestore), err
}

// List takes label and field selectors, and returns the list of restore objects that match those selectors.
func (c *FakeRestores) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.KuberlogicServiceRestoreList, err error) {
	obj, err := c.Fake.Invokes(testing.NewRootListAction(restoreResource, restoreKind, opts), &v1alpha1.KuberlogicServiceRestoreList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.KuberlogicServiceRestoreList{ListMeta: obj.(*v1alpha1.KuberlogicServiceRestoreList).ListMeta}
	for _, item := range obj.(*v1alpha1.KuberlogicServiceRestoreList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested restore objects.
func (c *FakeRestores) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.InvokesWatch(testing.NewRootWatchAction(restoreResource, opts))
}

// Create takes the representation of a restore object and creates it.  Returns the server's representation of the restore object, and an error, if there is any.
func (c *FakeRestores) Create(ctx context.Context, pod *v1alpha1.KuberlogicServiceRestore, opts v1.CreateOptions) (result *v1alpha1.KuberlogicServiceRestore, err error) {
	obj, err := c.Fake.Invokes(testing.NewRootCreateAction(restoreResource, pod), &v1alpha1.KuberlogicServiceRestore{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KuberlogicServiceRestore), err
}

// Delete takes name of the restore object and deletes it. Returns an error if one occurs.
func (c *FakeRestores) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.Invokes(testing.NewRootDeleteAction(restoreResource, name), &v1alpha1.KuberlogicServiceRestore{})
	return err
}

// Patch applies the patch and returns the patched restore object.
func (c *FakeRestores) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.KuberlogicServiceRestore, err error) {
	obj, err := c.Fake.Invokes(testing.NewRootPatchSubresourceAction(restoreResource, name, pt, data, subresources...), &v1alpha1.KuberlogicServiceRestore{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KuberlogicServiceRestore), err
}
