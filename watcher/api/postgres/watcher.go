package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	kuberlogicv1 "github.com/kuberlogic/operator/api/v1"
	util "github.com/kuberlogic/operator/util/kuberlogic"
	"github.com/kuberlogic/operator/watcher/api/base"
	"github.com/kuberlogic/operator/watcher/api/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

type Session struct {
	base.BaseSession
}

func New(cm *kuberlogicv1.KuberLogicService, client *kubernetes.Clientset, db, table string) (*Session, error) {
	session := &Session{}

	session.ClusterType = cm.Spec.Type
	session.ClusterNamespace = cm.Namespace
	session.Database = db
	session.Table = table
	session.Port = 5432

	if _, _, secret, err := util.GetClusterCredentialsInfo(cm); err != nil {
		return nil, err
	} else {
		session.ClusterCredentialsSecret = secret
	}

	if name, err := util.GetClusterName(cm); err != nil {
		return nil, err
	} else {
		session.ClusterName = name
	}
	if err := session.SetMaster(client); err != nil {
		return nil, err
	}
	if err := session.SetReplicas(client); err != nil {
		return nil, err
	}
	if err := session.SetCredentials(client); err != nil {
		return nil, err
	}
	return session, nil
}

func (session *Session) GetDatabase() common.Database {
	return &Database{session}
}

func (session *Session) GetUser() common.User {
	return &User{session}
}

func (session *Session) SetMaster(client *kubernetes.Clientset) error {
	pods, err := session.GetPods(client, client2.MatchingLabels{
		"application":  "spilo",
		"cluster-name": session.ClusterName,
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

	session.MasterIP = pods.Items[0].Status.PodIP

	return nil
}

func (session *Session) SetReplicas(client *kubernetes.Clientset) error {
	pods, err := session.GetPods(client, client2.MatchingLabels{
		"application":  "spilo",
		"cluster-name": session.ClusterName,
		"spilo-role":   "replica",
	})
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		session.ReplicaIPs = append(session.ReplicaIPs, pod.Status.PodIP)
	}
	return nil
}

func (session *Session) SetCredentials(client *kubernetes.Clientset) error {
	secret, err := client.CoreV1().Secrets(session.ClusterNamespace).Get(context.TODO(), session.ClusterCredentialsSecret, metav1.GetOptions{})
	if err != nil {
		return err
	}
	session.Password = string(secret.Data["password"])
	session.Username = string(secret.Data["username"])
	return nil
}

//////

func dbAlreadyExists(err error) bool {
	return strings.Contains(err.Error(), "already exists")
}

func (session *Session) SetupDDL() error {
	if err := session.GetDatabase().Create(session.Database); err != nil && !dbAlreadyExists(err) {
		return err
	}
	if err := session.CreateTable(); err != nil {
		return err
	}
	return nil
}

func (session *Session) ConnectionString(host, db string) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		session.Username, session.Password, host, session.Port, db)
}

func (session *Session) CreateTable() error {
	ctx := context.TODO()
	conn, err := pgx.Connect(ctx, session.ConnectionString(session.MasterIP, session.Database))
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx,
		"select table_name from information_schema.tables where table_name=$1", session.Table)
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
		if existingTable == session.Table {
			return nil
		}
	}
	_, err = conn.Exec(ctx,
		fmt.Sprintf("create table %s(id serial primary key)", session.Table))
	return err
}

func (session *Session) ReadLastRecord(host, table string, duration int64) (int, error) {
	config, err := pgx.ParseConfig(session.ConnectionString(host, session.Database))
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

func (session *Session) WriteRecord(host, table string, duration int64) (int, error) {
	config, err := pgx.ParseConfig(session.ConnectionString(host, session.Database))
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

func (session *Session) RunQueries(delay common.Delay, duration common.Duration) {
	items := []base.Argument{
		{
			"read",
			"master",
			session.ReadLastRecord,
			session.MasterIP,
			delay.MasterRead,
			duration.MasterRead,
		},
		{
			"write",
			"master",
			session.WriteRecord,
			session.MasterIP,
			delay.MasterWrite,
			duration.MasterWrite,
		},
	}
	for i, ip := range session.ReplicaIPs {
		items = append(items, base.Argument{
			"read",
			fmt.Sprintf("replica-%d", i),
			session.ReadLastRecord,
			ip,
			delay.ReplicaRead,
			duration.ReplicaRead,
		})
	}

	for _, arg := range items {
		go session.MakeQuery(arg)
	}
}
