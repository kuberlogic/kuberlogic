package postgresql

import (
	"context"
	"fmt"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/service-operator/base"
	"github.com/kuberlogic/operator/modules/operator/service-operator/interfaces"
	util "github.com/kuberlogic/operator/modules/operator/service-operator/util/kuberlogic"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"
)

type Session struct {
	base.BaseSession
	//Cluster *Postgres
}

func NewSession(cm *kuberlogicv1.KuberLogicService, client *kubernetes.Clientset, db string) (*Session, error) {
	session := &Session{}

	session.Database = db
	session.Port = 5432
	session.ClusterNamespace = cm.Namespace

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

func (session *Session) GetDatabase() interfaces.Database {
	return &Database{session}
}

func (session *Session) GetUser() interfaces.User {
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

func (session *Session) ConnectionString(host, db string) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		session.Username, session.Password, host, session.Port, db)
}
