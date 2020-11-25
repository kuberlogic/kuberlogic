package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	"gitlab.com/cloudmanaged/operator/watcher/api/base"
	"gitlab.com/cloudmanaged/operator/watcher/api/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

type Watcher struct {
	base.Watcher
}

func New(cm *cloudlinuxv1.CloudManaged, client *kubernetes.Clientset, db, table string) (*Watcher, error) {
	w := &Watcher{}

	w.Cluster = cm
	w.Database = db
	w.Table = table
	w.Port = 5432

	if err := w.SetMaster(client); err != nil {
		return nil, err
	}
	if err := w.SetReplicas(client); err != nil {
		return nil, err
	}
	if err := w.SetCredentials(client); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *Watcher) SetMaster(client *kubernetes.Clientset) error {
	pods, err := w.GetPods(client, client2.MatchingLabels{
		"application":  "spilo",
		"cluster-name": w.Cluster.Name,
		"spilo-role":   "master",
	})
	if err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		return errors.New("master pod is not found")
	} else if len(pods.Items) > 1 {
		return errors.New("master pod is not unique in the cluster")
	}

	w.MasterIP = pods.Items[0].Status.PodIP

	return nil
}

func (w *Watcher) SetReplicas(client *kubernetes.Clientset) error {
	pods, err := w.GetPods(client, client2.MatchingLabels{
		"application":  "spilo",
		"cluster-name": w.Cluster.Name,
		"spilo-role":   "replica",
	})
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		w.ReplicaIPs = append(w.ReplicaIPs, pod.Status.PodIP)
	}
	return nil
}

func (w *Watcher) SetCredentials(client *kubernetes.Clientset) error {
	secrets, err := client.CoreV1().Secrets("").
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, secret := range secrets.Items {
		if secret.Name == fmt.Sprintf("postgres.%s.credentials", w.Cluster.Name) {
			w.Password = string(secret.Data["password"])
			w.Username = string(secret.Data["username"])
			break
		}
	}
	return nil
}

//////

func dbAlreadyExists(err error) bool {
	return strings.Contains(err.Error(), "already exists")
}

func (w *Watcher) SetupDDL() error {
	if err := w.CreateDatabase(); err != nil && !dbAlreadyExists(err) {
		return err
	}
	if err := w.CreateTable(); err != nil {
		return err
	}
	return nil
}

func (w *Watcher) ConnectionString(host, db string) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		w.Username, w.Password, host, w.Port, db)
}

func (w *Watcher) CreateDatabase() error {
	ctx := context.TODO()
	conn, err := pgx.Connect(ctx, w.ConnectionString(w.MasterIP, "postgres"))
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	query := fmt.Sprintf("CREATE DATABASE %s;", w.Database)

	_, err = conn.Exec(ctx, query)
	return err
}

func (w *Watcher) CreateTable() error {
	ctx := context.TODO()
	conn, err := pgx.Connect(ctx, w.ConnectionString(w.MasterIP, w.Database))
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx,
		"select table_name from information_schema.tables where table_name=$1", w.Table)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var existingTable string
		err = rows.Scan(&existingTable)
		if err != nil {
			return err
		}
		if existingTable == w.Table {
			return nil
		}
	}
	_, err = conn.Exec(ctx,
		fmt.Sprintf("create table %s(id serial primary key)", w.Table))
	return err
}

func (w *Watcher) ReadLastRecord(host, table string, duration int64) (int, error) {
	config, err := pgx.ParseConfig(w.ConnectionString(host, w.Database))
	if err != nil {
		return 0, err
	}
	config.PreferSimpleProtocol = true

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(duration)*time.Second+5*time.Second)
	defer cancel()

	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return 0, err
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx,
		// TODO: Potentially might be SQL injection,
		// need to figure out how to parametrize table names
		fmt.Sprintf("select id from %s order by id desc limit 1", table))
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if duration > 0 {
		time.Sleep(time.Second * time.Duration(duration))
	}

	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			return 0, err
		}
		return id, nil
	}
	return 0, nil // 0 rows in dataset
}

func (w *Watcher) WriteRecord(host, table string, duration int64) (int, error) {
	config, err := pgx.ParseConfig(w.ConnectionString(host, w.Database))
	if err != nil {
		return 0, err
	}
	// due to error: "prepared statement "lrupsc_6_0" already exists (SQLSTATE 42P05)"
	// need to switch over simple protocol
	// more - https://godoc.org/github.com/jackc/pgx#hdr-PgBouncer
	config.PreferSimpleProtocol = true

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(duration)*time.Second+5*time.Second)
	defer cancel()

	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return 0, err
	}
	defer conn.Close(ctx)

	var lastInsertId int
	err = conn.QueryRow(ctx,
		fmt.Sprintf("insert into %s values (default) returning id", table),
	).Scan(&lastInsertId)

	if duration > 0 {
		time.Sleep(time.Second * time.Duration(duration))
	}

	if err != nil {
		return 0, err
	}
	return lastInsertId, nil

}

func (w *Watcher) RunQueries(delay common.Delay, duration common.Duration) {
	items := []base.Argument{
		{
			"read",
			"master",
			w.ReadLastRecord,
			w.MasterIP,
			delay.MasterRead,
			duration.MasterRead,
		},
		{
			"write",
			"master",
			w.WriteRecord,
			w.MasterIP,
			delay.MasterWrite,
			duration.MasterWrite,
		},
	}
	for i, ip := range w.ReplicaIPs {
		items = append(items, base.Argument{
			"read",
			fmt.Sprintf("replica-%d", i),
			w.ReadLastRecord,
			ip,
			delay.ReplicaRead,
			duration.ReplicaRead,
		})
	}

	for _, arg := range items {
		go w.MakeQuery(arg)
	}
}
