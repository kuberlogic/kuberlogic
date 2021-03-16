package tests

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

type tUser struct {
	service tService
	name string
}

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

func (u *tUser) Create(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.setRequestBody(fmt.Sprintf(`     {
        "name": "%s",
        "password": "secret-password"
     }`, u.name))
	api.sendRequestTo(http.MethodPost, fmt.Sprintf("/services/%s:%s/users/", u.service.ns, u.service.name))
	api.responseCodeShouldBe(201)
}

func (u *tUser) CreateTheSameName(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.setRequestBody(fmt.Sprintf(`     {
        "name": "%s",
        "password": "secret-password"
     }`, u.name))
		api.sendRequestTo(http.MethodPost, fmt.Sprintf("/services/%s:%s/users/", u.service.ns, u.service.name))
		api.responseCodeShouldBe(400)
		//api.encodeResponseToJson()
		// different response for mysql & pg
		//api.fieldContains("message", fmt.Sprintf("role \"%s\" already exists", user))
}

func (u *tUser) List(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/users/", u.service.ns, u.service.name))
	api.responseCodeShouldBe(200)
	api.encodeResponseToJson()
	api.responseTypeOf(reflect.Slice)
	api.responseShouldMatchJson(fmt.Sprintf(`
     [
		{"name": "%s"}
     ]`, u.name))
}

func (u *tUser) UserNotFound(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/%s/",
		u.service.ns, u.service.name, u.name))
	api.responseCodeShouldBe(404)
}

func (u *tUser) Delete(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.setRequestBody(fmt.Sprintf(`     {
        "name": "%s"
     }`, u.name))
	api.sendRequestTo(http.MethodDelete, fmt.Sprintf("/services/%s:%s/users/%s",
		u.service.ns, u.service.name, u.name))
	api.responseCodeShouldBe(200)
}

func (u *tUser) ChangePassword(t *testing.T)  {
	api := newApi(t)
	api.setBearerToken()
	api.setRequestBody(fmt.Sprintf(`     {
        "name": "%s",
		"password": "new-secret-password"
     }`, u.name))
	api.sendRequestTo(http.MethodPut, fmt.Sprintf("/services/%s:%s/users/%s",
		u.service.ns, u.service.name, u.name))
	api.responseCodeShouldBe(200)
}

func makeTestUser(tu tUser) func(t *testing.T) {
	return func(t *testing.T) {
		steps := []func(t *testing.T){
			tu.service.Create,
			tu.service.WaitForStatus("Ready", 5, 2*60),
			tu.Create,
			tu.CreateTheSameName,
			tu.List,
			tu.UserNotFound,
			// TODO: make test connection with password
			tu.ChangePassword,
			// TODO: make test connection with another password
			tu.Delete,
			tu.service.Delete,
		}
		for _, item := range steps {
			t.Run(GetFunctionName(item), item)
		}
	}
}

func TestUser(t *testing.T) {
	for _, svc := range []tUser{
		{
			service: pgTestService,
			name: "foo",
		}, {
			service: mysqlTestService,
			name: "foo",
		}} {
		t.Run(svc.service.type_, makeTestUser(svc))
	}
}
