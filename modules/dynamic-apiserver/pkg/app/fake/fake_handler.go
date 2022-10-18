/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package fake

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/watch"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/testing"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/api/fake"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/app"
)

type FakeHandlers struct {
	testing.Fake
	app.Handlers
	discovery *fakediscovery.FakeDiscovery
	tracker   testing.ObjectTracker
}

func (h *FakeHandlers) Tracker() testing.ObjectTracker {
	return h.tracker
}

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)

	_ testing.FakeClient = &FakeHandlers{}
)

func (h *FakeHandlers) Services() app.ExtendedServiceInterface {
	s := &app.Services{}
	s.ServiceInterface = &fake.FakeServices{}
	return s
}

func NewSimpleClient(handlers app.Handlers, objects ...runtime.Object) *FakeHandlers {
	o := testing.NewObjectTracker(scheme, codecs.UniversalDecoder())
	for _, obj := range objects {
		if err := o.Add(obj); err != nil {
			panic(err)
		}
	}

	cs := &FakeHandlers{tracker: o}
	cs.Handlers = handlers
	cs.discovery = &fakediscovery.FakeDiscovery{Fake: &cs.Fake}
	cs.AddReactor("*", "*", testing.ObjectReaction(o))
	cs.AddWatchReactor("*", func(action testing.Action) (handled bool, ret watch.Interface, err error) {
		gvr := action.GetResource()
		ns := action.GetNamespace()
		w, err := o.Watch(gvr, ns)
		if err != nil {
			return false, nil, err
		}
		return true, w, nil
	})

	return cs
}
