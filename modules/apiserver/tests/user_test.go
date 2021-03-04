package tests

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

type TestUser struct{}

func TestUsersDoesNotAllowMethodDelete(t *testing.T) {
	api := newApi(t)
	api.sendRequestTo(http.MethodDelete, "/services/default:cloudmanaged-pg/users/")
	api.responseCodeShouldBe(http.StatusMethodNotAllowed)
	api.encodeResponseToJson()
	api.fieldContains("message", "method DELETE is not allowed")
}

func TestServiceNotFoundForUsers(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodGet, "/services/default:cloudmanaged-pg/users/")
	api.responseCodeShouldBe(400)
	api.encodeResponseToJson()
	api.responseTypeOf(reflect.Map)
	api.responseShouldMatchJson(`{"message":"kuberlogicservices.kuberlogic.com \"cloudmanaged-pg\" not found"}`)
}

func (u *TestUser) Create(ns, name, user string) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.setRequestBody(fmt.Sprintf(`     {
        "name": "%s",
        "password": "secret-password"
     }`, user))
		api.sendRequestTo(http.MethodPost, fmt.Sprintf("/services/%s:%s/users/", ns, name))
		api.responseCodeShouldBe(201)
	}
}

func (u *TestUser) CreateTheSameName(ns, name, user string) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.setRequestBody(fmt.Sprintf(`     {
        "name": "%s",
        "password": "secret-password"
     }`, user))
		api.sendRequestTo(http.MethodPost, fmt.Sprintf("/services/%s:%s/users/", ns, name))
		api.responseCodeShouldBe(400)
		//api.encodeResponseToJson()
		// different response for mysql & pg
		//api.fieldContains("message", fmt.Sprintf("role \"%s\" already exists", user))
	}
}

func (u *TestUser) List(ns, name, user string) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/users/", ns, name))
		api.responseCodeShouldBe(200)
		api.encodeResponseToJson()
		api.responseTypeOf(reflect.Slice)
		api.responseShouldMatchJson(fmt.Sprintf(`
     [
		{"name": "%s"}
     ]`, user))
	}
}

func (u *TestUser) UserNotFound(ns, name, user string) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/%s/", ns, name, user))
		api.responseCodeShouldBe(404)
	}
}

func (u *TestUser) Delete(ns, name, user string) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.setRequestBody(fmt.Sprintf(`     {
        "name": "%s"
     }`, user))
		api.sendRequestTo(http.MethodDelete, fmt.Sprintf("/services/%s:%s/users/%s", ns, name, user))
		api.responseCodeShouldBe(200)
	}
}

func (u *TestUser) ChangePassword(ns, name, user string) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.setRequestBody(fmt.Sprintf(`     {
        "name": "%s",
		"password": "new-secret-password"
     }`, user))
		api.sendRequestTo(http.MethodPut, fmt.Sprintf("/services/%s:%s/users/%s", ns, name, user))
		api.responseCodeShouldBe(200)
	}
}

func CreateUser(t *testing.T, ns, name, type_ string) {
	tu := TestUser{}
	ts := tService{ns: ns, name: name, type_: type_, force: false, replicas: 1}
	user := "foo"
	steps := []func(t *testing.T){
		ts.Create,
		ts.WaitForStatus("Ready", 5, 2*60),
		tu.Create(ns, name, user),
		tu.CreateTheSameName(ns, name, user),
		tu.List(ns, name, user),
		tu.UserNotFound(ns, name, user),
		// TODO: make test connection with password
		tu.ChangePassword(ns, name, user),
		// TODO: make test connection with another password
		tu.Delete(ns, name, user),
		ts.Delete,
	}

	for _, item := range steps {
		t.Run(GetFunctionName(item), item)
	}
}

func TestCreateUserPg(t *testing.T) {
	CreateUser(t, pgService.ns, pgService.name, pgService.type_)
}

func TestCreateUserMysql(t *testing.T) {
	CreateUser(t, mysqlService.ns, mysqlService.name, mysqlService.type_)
}
