/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package api

import (
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	utiltesting "k8s.io/client-go/util/testing"
	"net/http/httptest"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

type Test struct {
	*testing.T
	server  *httptest.Server
	handler *utiltesting.FakeHandler
}

func (t *Test) Close() {
	t.server.Close()
}

func (t *Test) Handler() *utiltesting.FakeHandler {
	return t.handler
}

func (t *Test) fakeClient(obj runtime.Object, status int) rest.Interface {
	body := runtime.EncodeOrDie(scheme.Codecs.LegacyCodec(v1alpha1.GroupVersion), obj)
	t.handler = &utiltesting.FakeHandler{
		StatusCode:   status,
		ResponseBody: body,
		T:            t,
	}

	t.server = httptest.NewServer(t.handler)
	c, err := restClient(t.server)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return c
}

func restClient(testServer *httptest.Server) (rest.Interface, error) {
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

func init() {
	_ = v1alpha1.AddToScheme(scheme.Scheme)
}
