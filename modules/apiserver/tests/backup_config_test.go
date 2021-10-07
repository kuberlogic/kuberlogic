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
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

type tBackupConfig struct {
	service tService
}

func TestServiceNotFoundForTestBackupConfig(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:simple-pg/backup-config", testNs))
	api.responseCodeShouldBe(400)
}

func (u *tBackupConfig) CreateWithoutSchedule(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.setRequestBody(`     {
        "enabled": true,
        "aws_access_key_id": "aws_access_key_id",
		"aws_secret_access_key": "aws_secret_access_key",
		"bucket": "bucket",
		"endpoint": "endpoint"
     }`)
	api.sendRequestTo(http.MethodPost, fmt.Sprintf("/services/%s:%s/backup-config",
		u.service.ns, u.service.name))
	api.responseCodeShouldBe(422)
}

func (u *tBackupConfig) Create(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.setRequestBody(`     {
        "enabled": true,
        "aws_access_key_id": "aws_access_key_id",
		"aws_secret_access_key": "aws_secret_access_key",
		"bucket": "bucket",
		"endpoint": "endpoint",
		"schedule": "* 1 * * *"
     }`)
	api.sendRequestTo(http.MethodPost, fmt.Sprintf("/services/%s:%s/backup-config",
		u.service.ns, u.service.name))
	api.responseCodeShouldBe(201)
}

func (u *tBackupConfig) Get(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/backup-config",
		u.service.ns, u.service.name))
	api.responseCodeShouldBe(200)
	api.encodeResponseToJson()
	api.responseTypeOf(reflect.Map)
	api.responseShouldMatchJson(`{
        "enabled": true,
        "aws_access_key_id": "aws_access_key_id",
		"aws_secret_access_key": "aws_secret_access_key",
		"bucket": "bucket",
		"endpoint": "endpoint",
		"schedule": "* 1 * * *"
     }`)
}

func (u *tBackupConfig) Delete(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodDelete, fmt.Sprintf("/services/%s:%s/backup-config",
		u.service.ns, u.service.name))
	api.responseCodeShouldBe(200)
}

func (u *tBackupConfig) ChangeConfig(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.setRequestBody(`     {
        "enabled": false,
        "aws_access_key_id": "key-secret",
		"aws_secret_access_key": "access-secret",
		"bucket": "changed-backup",
		"endpoint": "new-endpoint",
		"schedule": "* 2 * * *"
     }`)
	api.sendRequestTo(http.MethodPut, fmt.Sprintf("/services/%s:%s/backup-config",
		u.service.ns, u.service.name))
	api.responseCodeShouldBe(200)
}

func (u *tBackupConfig) GetChanged(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/backup-config",
		u.service.ns, u.service.name))
	api.responseCodeShouldBe(200)
	api.encodeResponseToJson()
	api.responseTypeOf(reflect.Map)
	api.responseShouldMatchJson(`{
        "enabled": false,
        "aws_access_key_id": "key-secret",
		"aws_secret_access_key": "access-secret",
		"bucket": "changed-backup",
		"endpoint": "new-endpoint",
		"schedule": "* 2 * * *"
     }`)
}

func makeTestBackupConfig(tbc tBackupConfig) func(t *testing.T) {
	return func(t *testing.T) {
		steps := []func(t *testing.T){
			tbc.service.Create,
			tbc.service.WaitForStatus("Ready", 5, 5*60),
			tbc.CreateWithoutSchedule,
			tbc.Create,
			tbc.Get,
			tbc.ChangeConfig,
			tbc.GetChanged,
			tbc.Delete,
			tbc.service.Delete,
		}

		for _, item := range steps {
			t.Run(GetFunctionName(item), item)
		}
	}
}

func TestBackupConfig(t *testing.T) {
	for _, svc := range []tBackupConfig{
		{
			service: pgTestService,
		}, {
			service: mysqlTestService,
		}} {
		t.Run(svc.service.type_, makeTestBackupConfig(svc))
	}
}
