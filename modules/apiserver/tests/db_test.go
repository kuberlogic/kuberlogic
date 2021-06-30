package tests

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

type tDb struct {
	service tService
	name    string
}

func TestDbDoesNotAllowMethodDelete(t *testing.T) {
	api := newApi(t)
	api.sendRequestTo(http.MethodDelete, fmt.Sprintf("/services/%s:cloudmanaged-pg/databases/", testNs))
	api.responseCodeShouldBe(http.StatusMethodNotAllowed)
	api.encodeResponseToJson()
	api.fieldContains("message", "method DELETE is not allowed")
}

func TestServiceNotFoundForDb(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:cloudmanaged-pg/databases/", testNs))
	api.responseCodeShouldBe(400)
	api.encodeResponseToJson()
	api.responseTypeOf(reflect.Map)
	api.responseShouldMatchJson(`{"message":"service not found"}`)
}

func (td *tDb) Create(t *testing.T) {

	api := newApi(t)
	api.setBearerToken()
	api.setRequestBody(fmt.Sprintf(`{
        "name": "%s"
     }`, td.name))
	api.sendRequestTo(http.MethodPost, fmt.Sprintf("/services/%s:%s/databases/",
		td.service.ns, td.service.name))

	if api.response.StatusCode != 201 {
		api.encodeResponseToJson()
		t.Logf("response: %v", api.jsonResponse)
	}
	api.responseCodeShouldBe(201)
}

func (td *tDb) CreateTheSameName(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.setRequestBody(fmt.Sprintf(`{
        "name": "%s"
     }`, td.name))
	api.sendRequestTo(http.MethodPost, fmt.Sprintf("/services/%s:%s/databases/",
		td.service.ns, td.service.name))
	api.responseCodeShouldBe(400)
}

func (td *tDb) EmptyList(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/databases/",
		td.service.ns, td.service.name))
	api.responseCodeShouldBe(200)
	api.encodeResponseToJson()
	api.responseTypeOf(reflect.Slice)
	api.responseShouldMatchJson("[]")

}

func (td *tDb) OneRecord(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/databases/", td.service.ns, td.service.name))
	api.responseCodeShouldBe(200)
	api.encodeResponseToJson()
	api.responseTypeOf(reflect.Slice)
	api.responseShouldMatchJson(fmt.Sprintf(`
     [
		{"name": "%s"}
     ]`, td.name))
}

func (td *tDb) GetMethodIsNotAllowed(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/databases/%s",
		td.service.ns, td.service.name, td.name))
	api.responseCodeShouldBe(405)
	api.encodeResponseToJson()
	api.responseTypeOf(reflect.Map)
	api.fieldContains("message", "method GET is not allowed")

}

func (td *tDb) Delete(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.setRequestBody(fmt.Sprintf(`{
        "name": "%s"
     }`, td.name))
	api.sendRequestTo(http.MethodDelete, fmt.Sprintf("/services/%s:%s/databases/%s",
		td.service.ns, td.service.name, td.name))
	api.responseCodeShouldBe(200)

}

func makeTestDb(td tDb) func(t *testing.T) {
	return func(t *testing.T) {
		steps := []func(t *testing.T){
			td.service.Create,
			td.service.WaitForStatus("Ready", 5, 5*60),

			td.Create,
			td.CreateTheSameName,
			td.OneRecord,
			td.GetMethodIsNotAllowed,
			// TODO: make test connection to database
			td.Delete,
			td.service.Delete,
		}

		for _, item := range steps {
			t.Run(GetFunctionName(item), item)
		}
	}
}

func TestDb(t *testing.T) {
	for _, svc := range []tDb{
		{
			service: pgTestService,
			name:    "foo",
		}, {
			service: mysqlTestService,
			name:    "foo",
		}} {
		t.Run(svc.service.type_, makeTestDb(svc))
	}
}
