package tests

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

type TestDb struct{}

func TestDbDoesNotAllowMethodDelete(t *testing.T) {
	api := newApi(t)
	api.sendRequestTo(http.MethodDelete, "/services/default:cloudmanaged-pg/databases/")
	api.responseCodeShouldBe(http.StatusMethodNotAllowed)
	api.encodeResponseToJson()
	api.fieldContains("message", "method DELETE is not allowed")
}

func TestServiceNotFoundForDb(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodGet, "/services/default:cloudmanaged-pg/databases/")
	api.responseCodeShouldBe(400)
	api.encodeResponseToJson()
	api.responseTypeOf(reflect.Map)
	api.responseShouldMatchJson(`{"message":"kuberlogicservices.kuberlogic.com \"cloudmanaged-pg\" not found"}`)
}

func (td *TestDb) Create(ns, name, db string) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.setRequestBody(fmt.Sprintf(`{
        "name": "%s"
     }`, db))
		api.sendRequestTo(http.MethodPost, fmt.Sprintf("/services/%s:%s/databases/", ns, name))
		api.responseCodeShouldBe(201)
	}
}

func (td *TestDb) CreateTheSameName(ns, name, db string) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.setRequestBody(fmt.Sprintf(`{
        "name": "%s"
     }`, db))
		api.sendRequestTo(http.MethodPost, fmt.Sprintf("/services/%s:%s/databases/", ns, name))
		api.responseCodeShouldBe(400)
	}
}

func (td *TestDb) EmptyList(ns, name string) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/databases/", ns, name))
		api.responseCodeShouldBe(200)
		api.encodeResponseToJson()
		api.responseTypeOf(reflect.Slice)
		api.responseShouldMatchJson("[]")
	}
}

func (td *TestDb) OneRecord(ns, name, db string) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/databases/", ns, name))
		api.responseCodeShouldBe(200)
		api.encodeResponseToJson()
		api.responseTypeOf(reflect.Slice)
		api.responseShouldMatchJson(fmt.Sprintf(`
     [
		{"name": "%s"}
     ]`, db))
	}
}

func (td *TestDb) GetMethodIsNotAllowed(ns, name, db string) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/databases/%s", ns, name, db))
		api.responseCodeShouldBe(405)
		api.encodeResponseToJson()
		api.responseTypeOf(reflect.Map)
		api.fieldContains("message", "method GET is not allowed")
	}
}

func (td *TestDb) Delete(ns, name, db string) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.setRequestBody(fmt.Sprintf(`{
        "name": "%s"
     }`, db))
		api.sendRequestTo(http.MethodDelete, fmt.Sprintf("/services/%s:%s/databases/%s", ns, name, db))
		api.responseCodeShouldBe(200)
	}
}

func CreateDb(t *testing.T, ns, name, type_ string) {
	td := TestDb{}
	ts := tService{ns: ns, name: name, type_: type_, force: false, replicas: 1}
	db := "foo"
	steps := []func(t *testing.T){
		ts.Create,
		ts.WaitForStatus("Ready", 5, 2*60),
		//wait(10 * 60),
		td.Create(ns, name, db),
		td.CreateTheSameName(ns, name, db),
		td.OneRecord(ns, name, db),
		td.GetMethodIsNotAllowed(ns, name, db),
		// TODO: make test connection to database
		td.Delete(ns, name, db),
		ts.Delete,
	}

	for _, item := range steps {
		t.Run(GetFunctionName(item), item)
	}

}

func TestCreateDbPg(t *testing.T) {
	CreateDb(t, pgService.ns, pgService.name, pgService.type_)
}

func TestCreateDbMysql(t *testing.T) {
	CreateDb(t, mysqlService.ns, mysqlService.name, mysqlService.type_)
}
