package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/go-openapi/runtime/middleware"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	clienttesting "k8s.io/client-go/testing"
	utiltesting "k8s.io/client-go/util/testing"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/api/fake"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

var (
	//scheme = runtime.NewScheme()
	//codecs = serializer.NewCodecFactory(scheme)

	_ clienttesting.FakeClient = &FakeHandlers{}
)

type TestLog struct {
	t *testing.T
}

type TestClient struct {
	client  *rest.RESTClient
	server  *httptest.Server
	handler *utiltesting.FakeHandler
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
	w.t.Log(payload, w.check)

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

func createTestClient(obj runtime.Object, status int, t *testing.T) TestClient {
	body := runtime.EncodeOrDie(scheme.Codecs.LegacyCodec(v1alpha1.GroupVersion), obj)
	handler := &utiltesting.FakeHandler{
		StatusCode:   status,
		ResponseBody: body,
		T:            t,
	}

	testServer := httptest.NewServer(handler)
	c, err := restClient(testServer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return TestClient{
		client:  c,
		server:  testServer,
		handler: handler,
	}
}

func restClient(testServer *httptest.Server) (*rest.RESTClient, error) {
	c, err := rest.RESTClientFor(&rest.Config{
		Host: testServer.URL,
		ContentConfig: rest.ContentConfig{
			GroupVersion:         &v1.SchemeGroupVersion,
			NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
		},
		Username: "user",
		Password: "pass",
	})
	return c, err
}

func setup() error {
	err := v1alpha1.AddToScheme(scheme.Scheme)
	return err
}

func tearDown() {}

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		fmt.Printf("setup test failure: %s\n", err)
		os.Exit(1)
	}
	code := m.Run()
	if code == 0 {
		// no need destroy if the tests are failed
		tearDown()
	}
	os.Exit(code)
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
	// can not using runtime.NewScheme()
	// need guaranty only tracking passing objects
	// i.e each test case should have its own scheme and unique objects
	//newScheme := runtime.NewScheme()
	//newCodec := serializer.NewCodecFactory(newScheme)
	//err := v1alpha1.AddToScheme(newScheme)
	//if err != nil {
	//	panic(err)
	//}

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
	baseHandlers := &handlers{
		log: &TestLog{t: t},
	}
	client := newSimpleClient(objects...)
	baseHandlers.services = client.Services() // override services with fake
	baseHandlers.backups = client.Backups()   // override backups with fake
	baseHandlers.restores = client.Restores() // override restores with fake
	client.Handlers = baseHandlers            // set actual handlers
	return client
}

type testCase struct {
	name    string
	objects []runtime.Object
	result  interface{}
	status  int
	params  interface{}
	helpers []func(args ...interface{}) error
}
