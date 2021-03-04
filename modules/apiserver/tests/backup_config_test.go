package tests

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

type TestBackupConfig struct{}

func TestServiceNotFoundForTestBackupConfig(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodGet, "/services/default:simple-pg/backup-config")
	api.responseCodeShouldBe(404)
}

func (u *TestBackupConfig) Create(ns, name string) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.setRequestBody(`     {
        "enabled": true,
        "aws_access_key_id": "aws_access_key_id",
		"aws_secret_access_key": "aws_secret_access_key",
		"bucket": "bucket",
		"endpoint": "endpoint"
     }`)
		api.sendRequestTo(http.MethodPost, fmt.Sprintf("/services/%s:%s/backup-config", ns, name))
		api.responseCodeShouldBe(201)
	}
}

func (u *TestBackupConfig) Get(ns, name string) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/backup-config", ns, name))
		api.responseCodeShouldBe(200)
		api.encodeResponseToJson()
		api.responseTypeOf(reflect.Map)
		api.responseShouldMatchJson(`{
        "enabled": true,
        "aws_access_key_id": "aws_access_key_id",
		"aws_secret_access_key": "aws_secret_access_key",
		"bucket": "bucket",
		"endpoint": "endpoint"
     }`)
	}
}

func (u *TestBackupConfig) Delete(ns, name string) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.sendRequestTo(http.MethodDelete, fmt.Sprintf("/services/%s:%s/backup-config", ns, name))
		api.responseCodeShouldBe(200)
	}
}

func (u *TestBackupConfig) ChangeConfig(ns, name string) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.setRequestBody(`     {
        "enabled": false,
        "aws_access_key_id": "key-secret",
		"aws_secret_access_key": "access-secret",
		"bucket": "changed-backup",
		"endpoint": "new-endpoint"
     }`)
		api.sendRequestTo(http.MethodPut, fmt.Sprintf("/services/%s:%s/backup-config", ns, name))
		api.responseCodeShouldBe(200)
	}
}

func (u *TestBackupConfig) GetChanged(ns, name string) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/backup-config", ns, name))
		api.responseCodeShouldBe(200)
		api.encodeResponseToJson()
		api.responseTypeOf(reflect.Map)
		api.responseShouldMatchJson(`{
        "enabled": false,
        "aws_access_key_id": "key-secret",
		"aws_secret_access_key": "access-secret",
		"bucket": "changed-backup",
		"endpoint": "new-endpoint"
     }`)
	}
}

func CreateBackupConfig(t *testing.T, ns, name, type_ string) {
	tbc := TestBackupConfig{}
	ts := tService{ns: ns, name: name, type_: type_, force: false, replicas: 1}
	steps := []func(t *testing.T){
		ts.Create,
		ts.WaitForStatus("Ready", 5, 2*60),
		tbc.Create(ns, name),
		tbc.Get(ns, name),
		tbc.ChangeConfig(ns, name),
		tbc.GetChanged(ns, name),
		tbc.Delete(ns, name),
		ts.Delete,
	}

	for _, item := range steps {
		t.Run(GetFunctionName(item), item)
	}
}

func TestCreateBackupConfigPg(t *testing.T) {
	CreateBackupConfig(t, pgService.ns, pgService.name, pgService.type_)
}

func TestCreateBackupConfigMysql(t *testing.T) {
	CreateBackupConfig(t, mysqlService.ns, mysqlService.name, mysqlService.type_)
}
