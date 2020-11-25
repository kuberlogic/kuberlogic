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

type Watcher struct {
	base.Watcher
}

func New(cm *cloudlinuxv1.CloudManaged, client *kubernetes.Clientset, db, table string) (*Watcher, error) {
	w := &Watcher{}

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

func (w *Watcher) SetCredentials(client *kubernetes.Clientset) error {
	secrets, err := client.CoreV1().Secrets("").
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, secret := range secrets.Items {
		if secret.Name == w.Cluster.Spec.SecretName {
			w.Password = string(secret.Data["ROOT_PASSWORD"])
			w.Username = "root"
			break
		}
	}
	return nil
}

func (w *Watcher) SetMaster(client *kubernetes.Clientset) error {
	pods, err := w.GetPods(client, client2.MatchingLabels{
		"mysql.presslabs.org/cluster": w.Cluster.Name,
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

	w.MasterIP = pods.Items[0].Status.PodIP

	return nil
}

func (w *Watcher) SetReplicas(client *kubernetes.Clientset) error {
	pods, err := w.GetPods(client, client2.MatchingLabels{
		"mysql.presslabs.org/cluster": w.Cluster.Name,
		"role":                        "replica",
		"healthy":                     "yes",
	})
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		w.ReplicaIPs = append(w.ReplicaIPs, pod.Status.PodIP)
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

func (w *Watcher) ConnectionString(host, db string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		w.Username, w.Password, host, w.Port, db)
}

func (w *Watcher) SetupDDL() error {
	if err := w.CreateDatabase(); err != nil && !dbAlreadyExists(err) {
		return err
	}
	if err := w.CreateTable(); err != nil && !tableAlreadyExists(err) {
		return err
	}
	return nil
}

func (w *Watcher) CreateDatabase() error {
	db, err := sql.Open("mysql", w.ConnectionString(w.MasterIP, ""))
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	query := fmt.Sprintf(`
	CREATE DATABASE %s;
	`, w.Database)

	_, err = db.Exec(query)
	return err
}

func (w *Watcher) CreateTable() error {
	db, err := sql.Open("mysql", w.ConnectionString(w.MasterIP, w.Database))
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	query := fmt.Sprintf(`
		CREATE TABLE %s(
		   id INT AUTO_INCREMENT PRIMARY KEY
		);
	`, w.Table)

	_, err = db.Exec(query)
	return err
}

func (w *Watcher) ReadLastRecord(host, table string, duration int64) (int, error) {
	db, err := sql.Open("mysql", w.ConnectionString(host, w.Database))
	if err != nil {
		panic(err.Error())
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

func (w *Watcher) WriteRecord(host, table string, duration int64) (int, error) {

	db, err := sql.Open("mysql", w.ConnectionString(host, w.Database))
	if err != nil {
		panic(err.Error())
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
