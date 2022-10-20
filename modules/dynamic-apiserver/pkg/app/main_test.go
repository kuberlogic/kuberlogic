package app

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/go-openapi/runtime/middleware"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes"
	fake2 "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	clienttesting "k8s.io/client-go/testing"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/api/fake"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/config"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

var _ clienttesting.FakeClient = &FakeHandlers{}

type TestLog struct {
	t *testing.T
}

type Writer struct {
	t              *testing.T
	expectedStatus int
}
type Producer struct {
	t     *testing.T
	check interface{}
}

func (w *Writer) Header() http.Header {
	return http.Header{}
}

func (w *Writer) Write(_ []byte) (int, error) {
	return 0, nil
}

func (w *Writer) WriteHeader(statusCode int) {
	if statusCode != w.expectedStatus {
		w.t.Error("status does not equal: actual vs expected", statusCode, w.expectedStatus)
	}
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func (w *Producer) Produce(_ io.Writer, payload interface{}) error {
	switch w.check.(type) {
	case func(interface{}):
		w.check.(func(interface{}))(payload)
	default:
		if !reflect.DeepEqual(payload, w.check) {
			w.t.Logf("type actual vs expected: %T %T", payload, w.check)
			w.t.Error("object does not equal: actual vs expected",
				prettyPrint(payload),
				prettyPrint(w.check))
		}
	}
	return nil
}

func checkResponse(responder middleware.Responder, t *testing.T, status int, payload interface{}) {
	responder.WriteResponse(&Writer{
		t:              t,
		expectedStatus: status,
	}, &Producer{
		t:     t,
		check: payload,
	})
}

func (tl *TestLog) Debugw(msg string, keysAndValues ...interface{}) {
	tl.t.Log("debug", msg, keysAndValues)
}

func (tl *TestLog) Infow(msg string, keysAndValues ...interface{}) {
	tl.t.Log("info", msg, keysAndValues)
}

func (tl *TestLog) Warnw(msg string, keysAndValues ...interface{}) {
	tl.t.Log("warn", msg, keysAndValues)
}

func (tl *TestLog) Errorw(msg string, keysAndValues ...interface{}) {
	tl.t.Log("error", msg, keysAndValues)
}

func (tl *TestLog) Fatalw(msg string, keysAndValues ...interface{}) {
	tl.t.Log("fatal", msg, keysAndValues)
}

func (tl *TestLog) Infof(format string, args ...interface{}) {
	tl.t.Logf(format, args...)
}

func (tl *TestLog) Sync() error {
	return nil
}

func init() {
	_ = v1alpha1.AddToScheme(scheme.Scheme)
}

//------------------------------------------------------------------------------
// mocking based on operation (get, list, create, delete, watch, patch, etc.) of services/backups/restores

type FakeHandlers struct {
	clienttesting.Fake
	Handlers
	discovery *fakediscovery.FakeDiscovery
	tracker   clienttesting.ObjectTracker
}

func (h *FakeHandlers) Tracker() clienttesting.ObjectTracker {
	return h.tracker
}

func (h *FakeHandlers) Services() ExtendedServiceInterface {
	s := &services{}
	s.ServiceInterface = &fake.FakeServices{Fake: &h.Fake}
	return s
}

func (h *FakeHandlers) Backups() ExtendedBackupInterface {
	b := &backups{}
	b.BackupInterface = &fake.FakeBackup{Fake: &h.Fake}
	return b
}

func (h *FakeHandlers) Restores() ExtendedRestoreInterface {
	r := &restores{}
	r.RestoreInterface = &fake.FakeRestores{Fake: &h.Fake}
	return r
}

func newSimpleClient(objects ...runtime.Object) *FakeHandlers {
	o := clienttesting.NewObjectTracker(scheme.Scheme, scheme.Codecs.UniversalDecoder())
	for _, obj := range objects {
		if err := o.Add(obj); err != nil {
			panic(err)
		}
	}

	h := &FakeHandlers{tracker: o}
	h.discovery = &fakediscovery.FakeDiscovery{Fake: &h.Fake}
	h.AddReactor("*", "*", clienttesting.ObjectReaction(o))
	h.AddWatchReactor("*", func(action clienttesting.Action) (handled bool, ret watch.Interface, err error) {
		gvr := action.GetResource()
		ns := action.GetNamespace()
		w, err := o.Watch(gvr, ns)
		if err != nil {
			return false, nil, err
		}
		return true, w, nil
	})
	return h
}

func newFakeHandlers(t *testing.T, objects ...runtime.Object) *FakeHandlers {
	internalObjects, customObjects := splitObjects(objects)
	clientset := fake2.NewSimpleClientset(internalObjects...)
	return newFakeHandlersWithClientset(t, clientset, customObjects...)
}

func newFakeHandlersWithClientset(t *testing.T, clientset kubernetes.Interface, objects ...runtime.Object) *FakeHandlers {
	baseHandlers := &handlers{
		log: &TestLog{t: t},
		config: &config.Config{
			Domain: "kuberlogic.local",
		},
		clientset: clientset,
	}
	client := newSimpleClient(objects...)
	baseHandlers.services = client.Services() // override services with fake
	baseHandlers.backups = client.Backups()   // override backups with fake
	baseHandlers.restores = client.Restores() // override restores with fake
	client.Handlers = baseHandlers            // set actual handlers
	return client
}

func splitObjects(objects []runtime.Object) ([]runtime.Object, []runtime.Object) {
	customObjects, internalObjects := make([]runtime.Object, 0), make([]runtime.Object, 0)
	for _, obj := range objects {
		switch obj.(type) {
		case *v1alpha1.KuberLogicService, *v1alpha1.KuberLogicServiceList,
			*v1alpha1.KuberlogicServiceBackup, *v1alpha1.KuberlogicServiceBackupList,
			*v1alpha1.KuberlogicServiceRestore, *v1alpha1.KuberlogicServiceRestoreList:
			customObjects = append(customObjects, obj)
		default:
			internalObjects = append(internalObjects, obj)
		}
	}
	return internalObjects, customObjects
}

type testCase struct {
	name    string
	objects []runtime.Object
	result  interface{}
	status  int
	params  interface{}
	helpers []func(args ...interface{}) error
}
