package app

import (
	"encoding/json"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	"io"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	utiltesting "k8s.io/client-go/util/testing"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
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
	switch w.check.(type) {
	case func(interface{}):
		w.check.(func(interface{}))(payload)
	default:
		if !reflect.DeepEqual(payload, w.check) {
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

func (tl *TestLog) Sync() error {
	return nil
}

func createTestClient(obj runtime.Object, status int, t *testing.T) TestClient {
	body := runtime.EncodeOrDie(scheme.Codecs.LegacyCodec(kuberlogicv1.GroupVersion), obj)
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
	err := kuberlogicv1.AddToScheme(scheme.Scheme)
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
