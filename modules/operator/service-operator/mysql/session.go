/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	base2 "github.com/kuberlogic/kuberlogic/modules/operator/service-operator/base"
	"github.com/kuberlogic/kuberlogic/modules/operator/service-operator/interfaces"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"
)

type Session struct {
	base2.BaseSession
	client kubernetes.Interface
}

func NewSession(op interfaces.OperatorInterface, cm *kuberlogicv1.KuberLogicService, client kubernetes.Interface, db string) (*Session, error) {
	session := &Session{
		client: client,
	}

	session.ClusterName = op.Name(cm)
	session.ClusterNamespace = cm.Namespace

	session.Database = db
	session.Port = 3306

	session.ClusterCredentialsSecret, session.PasswordField = op.GetInternalDetails().GetDefaultConnectionPassword()

	if err := session.fillMaster(); err != nil {
		return nil, err
	}
	if err := session.fillReplicas(); err != nil {
		return nil, err
	}
	if err := session.fillCredentials(); err != nil {
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

func (session *Session) fillCredentials() error {
	secret, err := session.client.CoreV1().Secrets(session.ClusterNamespace).Get(
		context.TODO(), session.ClusterCredentialsSecret, metav1.GetOptions{})

	if err != nil {
		return err
	}
	session.Password = string(secret.Data[passwordField])
	session.Username = masterUser
	return nil
}

func (session *Session) SetCredentials(password string) error {
	s := v1.Secret{
		StringData: map[string]string{
			passwordField: password,
		},
	}

	patch, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("error decode secret: %s", err)
	}

	secret, err := session.client.CoreV1().Secrets(
		session.ClusterNamespace,
	).Patch(
		context.TODO(),
		session.ClusterCredentialsSecret,
		types.MergePatchType,
		patch,
		metav1.PatchOptions{})
	if err != nil {
		return err
	}
	session.Password = string(secret.Data[passwordField])
	session.Username = masterUser

	return nil
}

func (session *Session) fillMaster() error {
	pods, err := session.GetPods(session.client, client2.MatchingLabels{
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

func (session *Session) fillReplicas() error {
	pods, err := session.GetPods(session.client, client2.MatchingLabels{
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

func (session *Session) CreateTable(table string) error {
	db, err := sql.Open("mysql", session.ConnectionString(session.MasterIP, session.Database))
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf(`
		CREATE TABLE %s(
		   id INT AUTO_INCREMENT PRIMARY KEY
		);
	`, table)

	_, err = db.Exec(query)
	return err
}
