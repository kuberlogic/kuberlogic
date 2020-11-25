package base

import (
	"bufio"
	"context"
	"fmt"
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"
	"path"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type Watcher struct {
	Cluster    *cloudlinuxv1.CloudManaged
	MasterIP   string
	ReplicaIPs []string

	Port     int
	Database string
	Table    string

	Username string
	Password string
}

type Argument struct {
	Operation string
	Name      string

	Func     func(host, table string, duration int64) (int, error)
	Host     string
	Delay    int64
	Duration int64
}

func (w *Watcher) GetPods(client *kubernetes.Clientset, matchingLabels client2.MatchingLabels) (*v1.PodList, error) {
	labelMap, err := metav1.LabelSelectorAsMap(&metav1.LabelSelector{
		MatchLabels: matchingLabels,
	})
	if err != nil {
		return nil, err
	}

	options := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelMap).String(),
	}

	pods, err := client.CoreV1().
		Pods("").
		List(context.TODO(), options)
	if err != nil {
		return nil, err
	}
	return pods, nil
}

func (w *Watcher) String() string {
	return fmt.Sprintf(`
Cluster: %s
Master IP: %s
Replica ips: %s
Username: %s
Password: %s
Database: %s
Table: %s
Port: %d`,
		w.Cluster.Spec.Type,
		w.MasterIP,
		w.ReplicaIPs,
		w.Username,
		w.Password,
		w.Database,
		w.Table,
		w.Port,
	)
}

func FormatTime() string {
	return time.Now().Format(time.Stamp)
}

func (w *Watcher) MakeQuery(arg Argument) {
	dir := fmt.Sprintf("/tmp/cm/%s/", w.Cluster.Name)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	name := fmt.Sprintf("%s-%s.log", arg.Name, arg.Operation)
	filename := path.Join(dir, name)
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(filename)
	writer := bufio.NewWriter(f)
	for {
		number, err := arg.Func(arg.Host, w.Table, arg.Duration)
		var s string
		if err != nil {
			s = fmt.Sprintf("%s: %s/%s: %v\n", FormatTime(), arg.Name, arg.Operation, err)
		} else {
			s = fmt.Sprintf("%s: %s/%s: %v\n", FormatTime(), arg.Name, arg.Operation, number)
		}
		_, err = writer.WriteString(s)
		if err != nil {
			log.Fatal(err)
		}
		err = writer.Flush()
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(time.Millisecond * time.Duration(arg.Delay))
	}
}
