package tests

import (
	"fmt"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	"github.com/kuberlogic/operator/modules/operator/service-operator/util/kuberlogic"
	"net/http"
	"os"
	"testing"
	"time"
)

type tBackupRestore struct {
	service      tService
	backupConfig tBackupConfig
	db           tDb

	accessKey string
	secretKey string
	bucket    string
	endpoint  string
	backup    *models.Backup

	table      string
	backupTime time.Time
}

func (tb *tBackupRestore) CreateSchedule(t *testing.T) {

	t.Logf("Current UTC time is %v", time.Now().UTC())
	t.Logf("Scheduled backup time is %v", tb.backupTime)

	api := newApi(t)
	api.setBearerToken()
	api.setRequestBody(fmt.Sprintf(`     {
        "enabled": true,
        "aws_access_key_id": "%s",
		"aws_secret_access_key": "%s",
		"bucket": "%s",
		"endpoint": "%s",
		"schedule": "%02d %02d * * *"
     }`, tb.accessKey, tb.secretKey, tb.bucket, tb.endpoint, tb.backupTime.Minute(), tb.backupTime.Hour()))
	api.sendRequestTo(http.MethodPost, fmt.Sprintf("/services/%s:%s/backup-config", tb.service.ns, tb.service.name))
	api.responseCodeShouldBe(201)

}

func (tb *tBackupRestore) EnsureClearConfig(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodDelete, fmt.Sprintf("/services/%s:%s/backup-config",
		tb.service.ns, tb.service.name))
}

func (tb *tBackupRestore) EnsureClearDB(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.setRequestBody(fmt.Sprintf(`{
        "name": "%s"
     }`, tb.db.name))
	api.sendRequestTo(http.MethodDelete, fmt.Sprintf("/services/%s:%s/databases/%s",
		tb.service.ns, tb.service.name, tb.db.name))
}

func (tb *tBackupRestore) GetLastBackup(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/backups",
		tb.service.ns, tb.service.name))
	api.responseCodeShouldBe(200)

	var backups []*models.Backup
	api.encodeResponseTo(&backups)
	if len(backups) == 0 {
		t.Errorf("backups is not found for %s:%s",
			tb.service.ns, tb.service.name)
		return
	}

	// Get the first one after backup time
	var found *models.Backup
	for _, item := range backups {
		if tb.backupTime.Before(time.Time(*item.LastModified)) {
			found = item
			break
		}
	}
	if found == nil {
		t.Errorf("backup is not found after %v", tb.backupTime)
	}
	t.Logf("backup is found %v", found)
	tb.backup = found
}

func (tb *tBackupRestore) RestoreFromBackup(t *testing.T) {

	if tb.backup == nil {
		t.Error("backup is not found")
		return
	}
	api := newApi(t)
	api.setBearerToken()
	api.setRequestBody(fmt.Sprintf(`     {
        "key": "%s",
        "database": "%s"
     }`, *tb.backup.File, tb.db.name))
	api.sendRequestTo(http.MethodPost, fmt.Sprintf("/services/%s:%s/restores", tb.service.ns, tb.service.name))
	api.responseCodeShouldBe(200)
}

func (tb *tBackupRestore) CheckSuccesfulRestore(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/restores", tb.service.ns, tb.service.name))
	api.responseCodeShouldBe(200)

	var restores []*models.Restore
	api.encodeResponseTo(&restores)

	var found *models.Restore
	for _, item := range restores {
		if *item.Status == "Success" {
			found = item
		}
	}
	if found == nil {
		t.Errorf("no succesful restore found for service %s:%s", tb.service.ns, tb.service.name)
	}
}

func (tb *tBackupRestore) CreateTable(t *testing.T) {
	client, resource, err := Connect(tb.service.ns, tb.service.name)
	if err != nil {
		t.Errorf("cannot connect to the k8s resource: %s", err)
		return
	}

	session, err := kuberlogic.GetSession(resource, client, tb.db.name)
	if err != nil {
		t.Errorf("cannot get session:%s", err)
		return
	}

	err = session.CreateTable(tb.table)
	if err != nil {
		t.Errorf("cannot create table: %s", err)
		return
	}
}

func makeTestBackupRestore(tb tBackupRestore) func(t *testing.T) {
	return func(t *testing.T) {
		now := time.Now().Local()

		tb.backupTime = time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			now.Hour(),
			now.Minute(),
			0, // set 0 for the seconds. Needed for the fast backups
			0,
			now.Location()).Add(2 * time.Minute).UTC() // 2 minutes delay needs for the resource creation

		steps := []func(t *testing.T){
			tb.service.Create,
			tb.service.WaitForStatus("Ready", 5, 5*60),
			tb.EnsureClearConfig,
			tb.EnsureClearDB,

			tb.db.EmptyList,
			tb.db.Create,
			tb.CreateTable, // mysql does not recover db if the entities  not exists
			tb.CreateSchedule,
			// TODO: waiting Success state of the backup resource
			wait(4 * 60),
			tb.db.OneRecord, // db exists
			tb.db.Delete,
			tb.db.EmptyList,

			tb.GetLastBackup,
			tb.RestoreFromBackup,
			// TODO: waiting Success state of the restore resource
			wait(2 * 60), // waiting for the restore from the backup
			tb.CheckSuccesfulRestore,
			tb.db.OneRecord,
			tb.db.Delete,
			tb.backupConfig.Delete, // deletes secret, so it breaks backups link
			tb.service.Delete,
		}

		for _, item := range steps {
			t.Run(GetFunctionName(item), item)
		}
	}
}

func TestBackupRestore(t *testing.T) {
	//t.FailNow()

	endpoint, exists := os.LookupEnv("MINIO_ENDPOINT")
	if !exists {
		t.Errorf("endpoint must be defined")
		return
	}
	accessKey, exists := os.LookupEnv("MINIO_ACCESS_KEY")
	if !exists {
		t.Errorf("accessKey must be defined")
		return
	}
	secretKey, exists := os.LookupEnv("MINIO_SECRET_KEY")
	if !exists {
		t.Errorf("secretKey must be defined")
		return
	}
	bucket, exists := os.LookupEnv("BUCKET")
	if !exists {
		t.Errorf("bucket must be defined")
		return
	}

	for _, svc := range []tBackupRestore{
		{
			service: pgTestService,
			backupConfig: tBackupConfig{
				service: pgTestService,
			},

			endpoint:  endpoint,
			accessKey: accessKey,
			secretKey: secretKey,
			bucket:    bucket,
			db: tDb{
				service: pgTestService,
				name:    "foo",
			},
			table: "foo",
		}, {
			service: mysqlTestService,
			backupConfig: tBackupConfig{
				service: mysqlTestService,
			},

			endpoint:  endpoint,
			accessKey: accessKey,
			secretKey: secretKey,
			bucket:    bucket,

			db: tDb{
				service: mysqlTestService,
				name:    "foo",
			},
			table: "foo",
		}} {
		//if svc.service.type_ == "mysql"{
		//	t.Skip("Temporary skipping. Fails with unknown reason on the github actions")
		//}
		t.Run(svc.service.type_, makeTestBackupRestore(svc))
	}
}
