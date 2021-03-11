package mysql

import (
	"context"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	base2 "github.com/kuberlogic/operator/modules/operator/service-operator/base"
	"github.com/kuberlogic/operator/modules/operator/service-operator/interfaces"
	util "github.com/kuberlogic/operator/modules/operator/service-operator/util/kuberlogic"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"
)

type Session struct {
	base2.BaseSession
}

func NewSession(cm *kuberlogicv1.KuberLogicService, client *kubernetes.Clientset, db string) (*Session, error) {
	w := &Session{}

	w.ClusterNamespace = cm.Namespace
	w.Database = db
	w.Port = 3306

	if _, _, secret, err := util.GetClusterCredentialsInfo(cm); err != nil {
		return nil, err
	} else {
		w.ClusterCredentialsSecret = secret
	}

	if name, err := util.GetClusterName(cm); err != nil {
		return nil, err
	} else {
		w.ClusterName = name
	}
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

func (session *Session) GetDatabase() interfaces.Database {
	return &Database{session}
}

func (session *Session) GetUser() interfaces.User {
	return &User{session}
}

func (session *Session) SetCredentials(client *kubernetes.Clientset) error {
	secret, err := client.CoreV1().Secrets(session.ClusterNamespace).Get(context.TODO(), session.ClusterCredentialsSecret, metav1.GetOptions{})

	if err != nil {
		return err
	}
	session.Password = string(secret.Data["ROOT_PASSWORD"])
	session.Username = "root"
	return nil
}

func (session *Session) SetMaster(client *kubernetes.Clientset) error {
	pods, err := session.GetPods(client, client2.MatchingLabels{
		"mysql.presslabs.org/cluster": session.ClusterName,
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
		"mysql.presslabs.org/cluster": session.ClusterName,
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

func (session *Session) ConnectionString(host, db string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		session.Username, session.Password, host, session.Port, db)
}
