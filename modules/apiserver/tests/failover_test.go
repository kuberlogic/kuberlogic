package tests

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/kuberlogic/operator/modules/operator/service-operator/util/kuberlogic"
	"io/ioutil"
	"os/exec"
	"testing"
)

type tFailover struct {
	service        tService
	backup         tBackupRestore
	pvcName        string
	masterPodName  string
	replicaPodName string
}

func (tf *tFailover) ExecCommand(name string, args ...string) func(t *testing.T) {
	return func(t *testing.T) {
		cmd := exec.Command(name, args...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			t.Error(err)
		}

		if err := cmd.Start(); err != nil {
			t.Error(err)
		}

		data, err := ioutil.ReadAll(stdout)
		if err != nil {
			t.Error(err)
		}

		err = cmd.Wait()
		t.Logf("%s %v: %s", name, args, string(data))
		if err != nil {
			t.Error(err)
		}
	}
}

func (tf tFailover) CheckPostgresqlCounter(expected int) func(t *testing.T) {
	return func(t *testing.T) {
		if tf.service.type_ != "postgresql" {
			t.Skipf("skipping, not a postgresql")
			return
		}

		client, resource, err := Connect(tf.service.ns, tf.service.name)
		if err != nil {
			t.Errorf("cannot connect to the k8s resource: %s", err)
			return
		}

		session, err := kuberlogic.GetSession(resource, client, tf.backup.db.name)
		if err != nil {
			t.Errorf("cannot get session:%s", err)
			return
		}

		replicas := session.GetReplicaIPs()
		if len(replicas) == 0 {
			t.Error("amount of replicas is zero")
			return
		}

		ctx := context.TODO()
		conn, err := pgx.Connect(ctx, session.ConnectionString(replicas[0], tf.backup.db.name))
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close(ctx)

		rows, err := conn.Query(ctx, fmt.Sprintf(`
SELECT count(id)
FROM %s;
`, tf.backup.table))
		if err != nil {
			t.Error(err)
			return
		}
		defer rows.Close()

		var count int
		rows.Next()
		err = rows.Scan(&count)
		if err != nil {
			t.Error(err)
			return
		}

		if count != expected {
			t.Errorf("count is mismatched: expected %d vs actual %d", expected, count)
		}
	}
}

func (tf tFailover) IncrementPostgresqlCounter(value int) func(t *testing.T) {
	return func(t *testing.T) {
		if tf.service.type_ != "postgresql" {
			t.Skipf("skipping, not a postgresql")
			return
		}

		client, resource, err := Connect(tf.service.ns, tf.service.name)
		if err != nil {
			t.Errorf("cannot connect to the k8s resource: %s", err)
			return
		}

		session, err := kuberlogic.GetSession(resource, client, tf.backup.db.name)
		if err != nil {
			t.Errorf("cannot get session:%s", err)
			return
		}

		config, err := pgx.ParseConfig(session.ConnectionString(session.GetMasterIP(), tf.backup.db.name))
		if err != nil {
			t.Error(err)
			return
		}
		// due to error: "prepared statement "lrupsc_6_0" already exists (SQLSTATE 42P05)"
		// need to switch over simple protocol
		// more - https://godoc.org/github.com/jackc/pgx#hdr-PgBouncer
		config.PreferSimpleProtocol = true

		ctx := context.TODO()
		conn, err := pgx.ConnectConfig(ctx, config)
		if err != nil {
			t.Error(err)
		}
		defer conn.Close(ctx)

		var lastInsertId int
		err = conn.QueryRow(ctx,
			fmt.Sprintf("insert into %s (id) values (%d) returning id", tf.backup.table, value),
		).Scan(&lastInsertId)
		if err != nil {
			t.Error(err)
		}
		if lastInsertId != value {
			t.Errorf("returned value is mismatched: exptected %d vs actual %d", value, lastInsertId)
		}
	}
}

func (tf tFailover) CheckMysqlCounter(value int) func(t *testing.T) {
	return func(t *testing.T) {
		if tf.service.type_ != "mysql" {
			t.Skipf("skipping, not a mysql")
			return
		}

		client, resource, err := Connect(tf.service.ns, tf.service.name)
		if err != nil {
			t.Errorf("cannot connect to the k8s resource: %s", err)
			return
		}

		session, err := kuberlogic.GetSession(resource, client, tf.backup.db.name)
		if err != nil {
			t.Errorf("cannot get session:%s", err)
			return
		}

		replicas := session.GetReplicaIPs()
		if len(replicas) == 0 {
			t.Error("amount of replicas is zero")
			return
		}

		db, err := sql.Open("mysql", session.ConnectionString(replicas[0], tf.backup.db.name))
		if err != nil {
			t.Error(err)
			return
		}
		defer db.Close()

		rows, err := db.Query(fmt.Sprintf("select count(id) from %s", tf.backup.table))
		if err != nil {
			t.Error(err)
			return
		}
		defer rows.Close()

		var count int
		rows.Next()
		err = rows.Scan(&count)
		if err != nil {
			t.Error(err)
			return
		}

		if count != value {
			t.Errorf("count is mismatched: expected %d vs actual %d", value, count)
		}
	}
}

func (tf tFailover) IncrementMysqlCounter(value int) func(t *testing.T) {
	return func(t *testing.T) {
		if tf.service.type_ != "mysql" {
			t.Skipf("skipping, not a mysql")
			return
		}

		client, resource, err := Connect(tf.service.ns, tf.service.name)
		if err != nil {
			t.Errorf("cannot connect to the k8s resource: %s", err)
			return
		}

		session, err := kuberlogic.GetSession(resource, client, tf.backup.db.name)
		if err != nil {
			t.Errorf("cannot get session:%s", err)
			return
		}

		db, err := sql.Open("mysql", session.ConnectionString(session.GetMasterIP(), tf.backup.db.name))
		if err != nil {
			t.Error(err)
		}
		defer db.Close()

		res, err := db.Exec(fmt.Sprintf("INSERT INTO %s (id) values (%d)", tf.backup.table, value))
		if err != nil {
			t.Error(err)
		}

		id, err := res.LastInsertId()
		if id != int64(value) {
			t.Errorf("returned value is mismatched: exptected %d vs actual %d", value, id)
		}
	}
}

func (tf *tFailover) RemovePersistentVolumeClaim(t *testing.T) {
	if tf.service.type_ == "mysql" {
		// when pvc is deleted and pod is recreated sometimes
		// https://github.com/presslabs/mysql-operator/issues/401
		t.Skipf("skipping, not a mysql")
		return
	}

	// kill the pvc, we'll check the failover and replication
	tf.ExecCommand("kubectl", "delete", "pvc", tf.pvcName, "--wait=false")(t)

	// kill the master replica by masterPodName
	tf.ExecCommand("kubectl", "delete", "pod", tf.masterPodName)(t)

	// wait Pending state, pod can not be Ready due to "pvc not found"
	wait(30)(t)
}

func makeTestFailover(tf tFailover) func(t *testing.T) {
	return func(t *testing.T) {
		steps := []func(t *testing.T){
			tf.service.Create,
			tf.service.WaitForStatus("Ready", 5, 2*60),

			tf.service.EditReplicas, // increase replicas to 2
			tf.service.WaitForStatus("Ready", 5, 5*60),

			// wait replica pod
			tf.service.WaitForRole("replica", "Running", 5, 3*60),

			// check replica & master before the failover
			tf.service.CheckRole(tf.masterPodName, "master", "Running"),
			tf.service.CheckRole(tf.replicaPodName, "replica", "Running"),

			// "Create db" endpoint returned 400 without timeout
			// dial tcp 172.17.0.15:3306: connect: connection refused
			wait(30),

			tf.backup.db.Create,
			tf.backup.CreateTable,

			tf.IncrementPostgresqlCounter(1),
			tf.IncrementMysqlCounter(1),

			// wait for synchronization with replicas
			wait(30),

			// check the counter
			tf.CheckPostgresqlCounter(1),
			tf.CheckMysqlCounter(1),

			tf.RemovePersistentVolumeClaim, // skipped for the mysql

			tf.ExecCommand("kubectl", "delete", "pod", tf.masterPodName),

			// wait for the failover
			wait(30),

			// now insert should work
			tf.IncrementPostgresqlCounter(2),
			tf.IncrementMysqlCounter(2),

			tf.service.WaitForStatus("Ready", 5, 5*60),
			tf.service.WaitForRole("replica", "Running", 5, 3*60),

			// now master & replica pod should be switched
			tf.service.CheckRole(tf.replicaPodName, "master", "Running"),
			tf.service.CheckRole(tf.masterPodName, "replica", "Running"),

			tf.CheckPostgresqlCounter(2),
			tf.CheckMysqlCounter(2),

			tf.service.Delete,
		}

		for _, item := range steps {
			t.Run(GetFunctionName(item), item)
		}
	}
}

func TestFailover(t *testing.T) {
	for _, svc := range []tFailover{
		{
			service: pgTestService,
			backup: tBackupRestore{
				service: pgTestService,

				db: tDb{
					service: pgTestService,
					name:    "failover",
				},
				table: "failover",
			},
			pvcName:        "pgdata-kuberlogic-pgsql-0",
			masterPodName:  "kuberlogic-pgsql-0",
			replicaPodName: "kuberlogic-pgsql-1",
		}, {
			service: mysqlTestService,
			backup: tBackupRestore{
				service: mysqlTestService,

				db: tDb{
					service: mysqlTestService,
					name:    "failover",
				},
				table: "failover",
			},
			pvcName:        "data-my-mysql-0",
			masterPodName:  "my-mysql-0",
			replicaPodName: "my-mysql-1",
		}} {
		t.Run(svc.service.type_, makeTestFailover(svc))
	}
}
