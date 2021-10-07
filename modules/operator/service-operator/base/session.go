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

package base

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"
)

type BaseSession struct {
	ClusterName      string
	ClusterNamespace string

	Database string
	Port     int
	Username string
	Password string

	ClusterCredentialsSecret string
	PasswordField            string

	MasterIP   string
	ReplicaIPs []string
}

func (session *BaseSession) GetPods(client *kubernetes.Clientset, matchingLabels client2.MatchingLabels) (*v1.PodList, error) {
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

func (session *BaseSession) GetMasterIP() string {
	return session.MasterIP
}

func (session *BaseSession) GetReplicaIPs() []string {
	return session.ReplicaIPs
}
