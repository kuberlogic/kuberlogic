package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	"gitlab.com/cloudmanaged/operator/watcher/api/base"
	"gitlab.com/cloudmanaged/operator/watcher/api/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

type Session struct {
	base.BaseSession
}

func New(cm *cloudlinuxv1.CloudManaged, client *kubernetes.Clientset, db, table string) (*Session, error) {
	w := &Session{}

	w.Cluster = cm
	w.Database = db
	w.Table = table
	w.Port = 3306

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

func (session *Session) GetDatabase() common.Database {
	return &Database{session}
}

func (session *Session) GetUser() common.User {
	return &User{session}
}

func (session *Session) SetCredentials(client *kubernetes.Clientset) error {
	secrets, err := client.CoreV1().Secrets("").
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, secret := range secrets.Items {
		if secret.Name == session.Cluster.Spec.SecretName {
			session.Password = string(secret.Data["ROOT_PASSWORD"])
			session.Username = "root"
			break
		}
	}
	return nil
}

func (session *Session) SetMaster(client *kubernetes.Clientset) error {
	pods, err := session.GetPods(client, client2.MatchingLabels{
		"mysql.presslabs.org/cluster": session.Cluster.Name,
		"role":                        "master",
		"healthy":                     "yes",
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
		"mysql.presslabs.org/cluster": session.Cluster.Name,
		"role":                        "replica",
		"healthy":                     "yes",
	})
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		session.ReplicaIPs = append(session.ReplicaIPs, pod.Status.PodIP)
	}
	return nil
}

/////

func dbAlreadyExists(err error) bool {
	return strings.Contains(err.Error(), "database exists")
}

func tableAlreadyExists(err error) bool {
	return strings.Contains(err.Error(), "already exists")
}

func (session *Session) ConnectionString(host, db string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		session.Username, session.Password, host, session.Port, db)
}

func (session *Session) SetupDDL() error {
	if err := session.GetDatabase().Create(session.Database); err != nil && !dbAlreadyExists(err) {
		return err
	}
	if err := session.CreateTable(); err != nil && !tableAlreadyExists(err) {
		return err
	}
	return nil
}

func (session *Session) CreateTable() error {
	db, err := sql.Open("mysql", session.ConnectionString(session.MasterIP, session.Database))
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf(`
		CREATE TABLE %s(
		   id INT AUTO_INCREMENT PRIMARY KEY
		);
	`, session.Table)

	_, err = db.Exec(query)
	return err
}

func (session *Session) ReadLastRecord(host, table string, duration int64) (int, error) {
	db, err := sql.Open("mysql", session.ConnectionString(host, session.Database))
	if err != nil {
		return 0, err
	}
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("select id from %s order by id desc limit 1", table))
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

	db, err := sql.Open("mysql", session.ConnectionString(host, session.Database))
	if err != nil {
		return 0, err
	}
	defer db.Close()

	res, err := db.Exec(fmt.Sprintf("INSERT INTO %s values (DEFAULT)", table))
	if err != nil {
		return 0, err
	}

	if duration > 0 {
		time.Sleep(time.Second * time.Duration(duration))
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil

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
