/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package api

import (
	"context"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func TestRestore(t *testing.T) {
	tt := &Test{T: t}
	defer tt.Close()

	c := NewRestores(tt.fakeClient(&v1alpha1.KuberlogicServiceRestore{}, 200))

	_, err := c.Create(context.TODO(), &v1alpha1.KuberlogicServiceRestore{}, v1.CreateOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = c.Patch(context.TODO(), "test", types.JSONPatchType, []byte{}, v1.PatchOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = c.Delete(context.TODO(), "test", v1.DeleteOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = c.Get(context.TODO(), "test", v1.GetOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = c.List(context.TODO(), v1.ListOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = c.Watch(context.TODO(), v1.ListOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
