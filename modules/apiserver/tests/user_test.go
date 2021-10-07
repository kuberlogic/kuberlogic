/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tests

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v4"
	"net/http"
	"reflect"
	"testing"
)

type tUser struct {
	service     tService
	name        string
	password    string
	newPassword string
	db          tDb
	masterUser  string
}

func TestUsersDoesNotAllowMethodDelete(t *testing.T) {
	api := newApi(t)
	api.sendRequestTo(http.MethodDelete, fmt.Sprintf("/services/%s:cloudmanaged-pg/users/", testNs))
	api.responseCodeShouldBe(http.StatusMethodNotAllowed)
	api.encodeResponseToJson()
	api.fieldContains("message", "method DELETE is not allowed")
}

func TestServiceNotFoundForUsers(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:cloudmanaged-pg/users/", testNs))
	api.responseCodeShouldBe(400)
	api.encodeResponseToJson()
	api.responseTypeOf(reflect.Map)
	api.responseShouldMatchJson(`{"message":"service not found"}`)
}

func (u *tUser) Create(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.setRequestBody(fmt.Sprintf(`     {
        "name": "%s",
        "password": "%s"
     }`, u.name, u.password))
	api.sendRequestTo(http.MethodPost, fmt.Sprintf("/services/%s:%s/users/", u.service.ns, u.service.name))
	api.responseCodeShouldBe(201)
}

func (u *tUser) CreateTheSameName(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.setRequestBody(fmt.Sprintf(`     {
        "name": "%s",
        "password": "%s"
     }`, u.name, u.password))
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
		{"name": "%s"},
		{"name": "%s"}
     ]`, u.name, u.masterUser))
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

func (u *tUser) ChangePassword(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.setRequestBody(fmt.Sprintf(`     {
        "name": "%s",
		"password": "%s"
     }`, u.name, u.newPassword))
	api.sendRequestTo(http.MethodPut, fmt.Sprintf("/services/%s:%s/users/%s",
		u.service.ns, u.service.name, u.name))
	api.responseCodeShouldBe(200)
}

func (u *tUser) ChangeMasterPassword(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.setRequestBody(fmt.Sprintf(`     {
        "name": "%s",
		"password": "new-secret-password"
     }`, u.masterUser))
	api.sendRequestTo(http.MethodPut, fmt.Sprintf("/services/%s:%s/users/%s",
		u.service.ns, u.service.name, u.masterUser))
	api.responseCodeShouldBe(200)
}

func (u *tUser) CheckConnection(user, password string) func(t *testing.T) {
	return func(t *testing.T) {
		session, err := GetSession(u.service.ns, u.service.name, u.db.name)
		if err != nil {
			t.Errorf("cannot get session:%s", err)
			return
		}

		connectionString := session.ConnectionString(session.GetMasterIP(), u.db.name)
		if u.service.type_ == "postgresql" {
			if user != "" && password != "" {
				connectionString = fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
					user, password, session.GetMasterIP(), 5432, u.db.name)
			}

			ctx := context.TODO()
			conn, err := pgx.Connect(ctx, connectionString)
			if err != nil {
				t.Errorf("cannot connect to pg: %s", err)
				return
			}
			defer conn.Close(ctx)

			_, err = conn.Exec(ctx, "select 42;")
			if err != nil {
				t.Errorf("cannot execute the select: %s", err)
				return
			}
		} else if u.service.type_ == "mysql" {
			if user != "" && password != "" {
				connectionString = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
					user, password, session.GetMasterIP(), 3306, u.db.name)
			}

			conn, err := sql.Open("mysql", session.ConnectionString(session.GetMasterIP(), u.db.name))
			if err != nil {
				t.Errorf("cannot open connection: %s", err)
				return
			}
			defer conn.Close()
			// Open doesn't open a connection. Validate DSN data:
			if err = conn.Ping(); err != nil {
				t.Errorf("cannot ping: %s", err)
				return
			}

			_, err = conn.Exec("select 42;")
			if err != nil {
				t.Errorf("cannot execute the select: %s", err)
				return
			}
		} else {
			t.Errorf("unknown service: %s", u.service.type_)
		}

	}
}
func makeTestUser(tu tUser) func(t *testing.T) {
	return func(t *testing.T) {
		steps := []func(t *testing.T){
			tu.service.Create,
			tu.service.WaitForStatus("Ready", 5, 5*60),
			tu.db.Create, // create db -> need to testing connection for user

			tu.CheckConnection("", ""),
			tu.ChangeMasterPassword,
			tu.CheckConnection("", ""),

			tu.Create,
			tu.CreateTheSameName,
			tu.List,
			tu.UserNotFound,
			tu.CheckConnection(tu.name, tu.password),
			tu.ChangePassword,
			tu.CheckConnection(tu.name, tu.newPassword),
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
			service:     pgTestService,
			name:        "foo",
			password:    "secret-password",
			newPassword: "new-secret-password",
			masterUser:  "kuberlogic",
			db: tDb{
				service: pgTestService,
				name:    "foo",
			},
		}, {
			service:     mysqlTestService,
			name:        "foo",
			password:    "secret-password",
			newPassword: "new-secret-password",
			masterUser:  "root",
			db: tDb{
				service: mysqlTestService,
				name:    "foo",
			},
		}} {
		t.Run(svc.service.type_, makeTestUser(svc))
	}
}
