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

type TestBackup struct {
	accessKey string
	secretKey string
	bucket    string
	endpoint  string
	backup    *models.Backup
}

func (tb *TestBackup) CreateSchedule(ns, name string, backupTime time.Time) func(t *testing.T) {
	return func(t *testing.T) {
		t.Logf("Current time is %v", time.Now())
		t.Logf("Scheduled backup time is %v", backupTime)

		api := newApi(t)
		api.setBearerToken()
		api.setRequestBody(fmt.Sprintf(`     {
        "enabled": true,
        "aws_access_key_id": "%s",
		"aws_secret_access_key": "%s",
		"bucket": "%s",
		"endpoint": "%s",
		"schedule": "%02d %02d * * *"
     }`, tb.accessKey, tb.secretKey, tb.bucket, tb.endpoint, backupTime.Minute(), backupTime.Hour()))
		api.sendRequestTo(http.MethodPost, fmt.Sprintf("/services/%s:%s/backup-config", ns, name))
		api.responseCodeShouldBe(201)
	}
}

func (tb *TestBackup) RemoveSchedule(ns, name string) func(t *testing.T) {
	return func(t *testing.T) {}
}

func (tb *TestBackup) EnsureClearConfig(ns, name string) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.sendRequestTo(http.MethodDelete, fmt.Sprintf("/services/%s:%s/backup-config", ns, name))
	}
}

func (tb *TestBackup) EnsureClearDB(ns, name, db string) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.setRequestBody(fmt.Sprintf(`{
        "name": "%s"
     }`, db))
		api.sendRequestTo(http.MethodDelete, fmt.Sprintf("/services/%s:%s/databases/%s", ns, name, db))
	}
}

func (tb *TestBackup) GetLastBackup(ns, name string, backupTime time.Time) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/backups", ns, name))
		api.responseCodeShouldBe(200)

		var backups []*models.Backup
		api.encodeResponseTo(&backups)
		if len(backups) == 0 {
			t.Errorf("backups is not found for %s:%s", ns, name)
			return
		}

		// Get the first one after backup time
		var found *models.Backup
		for _, item := range backups {
			if backupTime.Before(time.Time(*item.LastModified)) {
				found = item
				break
			}
		}
		if found == nil {
			t.Errorf("backup is not found after %v", backupTime)
		}
		tb.backup = found
	}
}

func (tb *TestBackup) RestoreFromBackup(ns, name, db string) func(t *testing.T) {
	return func(t *testing.T) {
		if tb.backup == nil {
			t.Error("backup is not found")
			return
		}
		api := newApi(t)
		api.setBearerToken()
		api.setRequestBody(fmt.Sprintf(`     {
        "key": "%s",
        "database": "%s"
     }`, *tb.backup.Key, db))
		api.sendRequestTo(http.MethodPost, fmt.Sprintf("/services/%s:%s/restore", ns, name))
		api.responseCodeShouldBe(200)
	}
}

func (tb *TestBackup) CreateTable(ns, name, db, table string) func(t *testing.T) {
	return func(t *testing.T) {
		client, resource, err := Connect(ns, name)
		if err != nil {
			t.Errorf("cannot connect to the k8s resource: %s", err)
			return
		}

		session, err := kuberlogic.GetSession(resource, client, db)
		if err != nil {
			t.Errorf("cannot get session:%s", err)
			return
		}

		err = session.CreateTable(table)
		if err != nil {
			t.Errorf("cannot create table: %s", err)
			return
		}
	}
}

func CreateBackup(t *testing.T, ns, name, type_ string) {
	endpoint, exists := os.LookupEnv("MINIO_ENDPOINT")
	if !exists {
		t.Errorf("endpoint must be exists")
		return
	}
	accessKey, exists := os.LookupEnv("MINIO_ACCESS_KEY")
	if !exists {
		t.Errorf("accessKey must be exists")
		return
	}
	secretKey, exists := os.LookupEnv("MINIO_SECRET_KEY")
	if !exists {
		t.Errorf("secretKey must be exists")
		return
	}
	bucket, exists := os.LookupEnv("TEST_BUCKET")
	if !exists {
		t.Errorf("bucket must be exists")
		return
	}

	tb := TestBackup{
		accessKey: accessKey,
		secretKey: secretKey,
		bucket:    bucket, // should be created before tests
		endpoint:  endpoint,
	}
	ts := tService{ns: ns, name: name, type_: type_, force: false, replicas: 1}
	tbc := TestBackupConfig{}
	td := TestDb{}
	db := "foo"
	now := time.Now().Local()
	backupTime := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		0, // set 0 for the seconds. Needed for the fast backups
		0,
		now.Location()).Add(2 * time.Minute) // 2 minutes delay needs for the resource creation

	steps := []func(t *testing.T){
		ts.Create,
		ts.WaitForStatus("Ready", 5, 2*60),
		tb.EnsureClearConfig(ns, name),
		tb.EnsureClearDB(ns, name, db),

		td.EmptyList(ns, name),
		td.Create(ns, name, db),
		tb.CreateTable(ns, name, db, "foo"), // mysql does not recover db if the entities  not exists
		tb.CreateSchedule(ns, name, backupTime),
		// TODO: waiting Success state of the backup resource
		wait(4 * 60),               // waiting for the backup
		td.OneRecord(ns, name, db), // db exists
		td.Delete(ns, name, db),
		td.EmptyList(ns, name),

		tb.GetLastBackup(ns, name, backupTime),
		tb.RestoreFromBackup(ns, name, db),
		// TODO: waiting Success state of the restore resource
		wait(2 * 60), // waiting for the restore from the backup
		td.OneRecord(ns, name, db),
		td.Delete(ns, name, db),
		tbc.Delete(ns, name), // deletes secret, so it breaks backups link
		ts.Delete,
	}

	for _, item := range steps {
		t.Run(GetFunctionName(item), item)
	}
}

func TestCreateBackupPg(t *testing.T) {
	CreateBackup(t, pgService.ns, pgService.name, pgService.type_)
}

func TestCreateBackupMysql(t *testing.T) {
	CreateBackup(t, mysqlService.ns, mysqlService.name, mysqlService.type_)
}
