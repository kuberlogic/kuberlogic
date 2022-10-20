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

// FakeBackup implements BackupInterface
type FakeBackup struct {
	Fake *testing.Fake
}

var (
	backupResource = schema.GroupVersionResource{Group: "kuberlogic.com", Version: "v1alpha1", Resource: "kuberlogicservicebackups"}
	backupKind     = schema.GroupVersionKind{Group: "kuberlogic.com", Version: "v1alpha1", Kind: "KuberlogicServiceBackup"}

	_ api.BackupInterface = &FakeBackup{}
)

// Get takes name of the backup, and returns the corresponding backup object, and an error if there is any.
func (c *FakeBackup) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.KuberlogicServiceBackup, err error) {
	obj, err := c.Fake.Invokes(testing.NewRootGetAction(backupResource, name), &v1alpha1.KuberlogicServiceBackup{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KuberlogicServiceBackup), err
}

// List takes label and field selectors, and returns the list of Backups that match those selectors.
func (c *FakeBackup) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.KuberlogicServiceBackupList, err error) {
	obj, err := c.Fake.Invokes(testing.NewRootListAction(backupResource, backupKind, opts), &v1alpha1.KuberlogicServiceBackupList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.KuberlogicServiceBackupList{ListMeta: obj.(*v1alpha1.KuberlogicServiceBackupList).ListMeta}
	for _, item := range obj.(*v1alpha1.KuberlogicServiceBackupList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested backups.
func (c *FakeBackup) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.InvokesWatch(testing.NewRootWatchAction(backupResource, opts))
}

// Create takes the representation of a backup and creates it.  Returns the server's representation of the service, and an error, if there is any.
func (c *FakeBackup) Create(ctx context.Context, pod *v1alpha1.KuberlogicServiceBackup, opts v1.CreateOptions) (result *v1alpha1.KuberlogicServiceBackup, err error) {
	obj, err := c.Fake.Invokes(testing.NewRootCreateAction(backupResource, pod), &v1alpha1.KuberlogicServiceBackup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KuberlogicServiceBackup), err
}

// Delete takes name of the backup and deletes it. Returns an error if one occurs.
func (c *FakeBackup) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.Invokes(testing.NewRootDeleteAction(backupResource, name), &v1alpha1.KuberlogicServiceBackup{})
	return err
}

// Patch applies the patch and returns the patched backup.
func (c *FakeBackup) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.KuberlogicServiceBackup, err error) {
	obj, err := c.Fake.Invokes(testing.NewRootPatchSubresourceAction(backupResource, name, pt, data, subresources...), &v1alpha1.KuberlogicServiceBackup{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KuberlogicServiceBackup), err
}
