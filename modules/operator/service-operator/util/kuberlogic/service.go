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

package kuberlogic

import (
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	serviceOperator "github.com/kuberlogic/kuberlogic/modules/operator/service-operator"
	"github.com/kuberlogic/kuberlogic/modules/operator/service-operator/interfaces"
	"k8s.io/client-go/kubernetes"
)

func GetCluster(cm *kuberlogicv1.KuberLogicService) (op interfaces.OperatorInterface, err error) {
	op, err = serviceOperator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)
	return
}

func GetClusterPodLabels(cm *kuberlogicv1.KuberLogicService) (master map[string]string, replica map[string]string, err error) {
	op, err := GetCluster(cm)
	if err != nil {
		return
	}
	master, replica = op.GetInternalDetails().GetPodMasterSelector(), op.GetInternalDetails().GetPodReplicaSelector()

	return
}

func GetClusterServices(cm *kuberlogicv1.KuberLogicService) (master string, replica string, err error) {
	op, err := GetCluster(cm)
	if err != nil {
		return
	}
	master, replica = op.GetInternalDetails().GetMasterService(), op.GetInternalDetails().GetReplicaService()
	return
}

func GetClusterServicePort(cm *kuberlogicv1.KuberLogicService) (p int, err error) {
	op, err := GetCluster(cm)
	if err != nil {
		return
	}
	p = op.GetInternalDetails().GetAccessPort()
	return
}

func GetClusterMainContainer(cm *kuberlogicv1.KuberLogicService) (c string, err error) {
	op, err := GetCluster(cm)
	if err != nil {
		return
	}
	c = op.GetInternalDetails().GetMainPodContainer()
	return
}

func GetClusterCredentialsInfo(cm *kuberlogicv1.KuberLogicService) (username, passwordField, secretName string, err error) {
	op, err := GetCluster(cm)
	if err != nil {
		return
	}
	details := op.GetInternalDetails()
	secretName, passwordField = details.GetDefaultConnectionPassword()
	username = details.GetDefaultConnectionUser()
	return
}

func GetSession(cm *kuberlogicv1.KuberLogicService, client *kubernetes.Clientset, db string) (session interfaces.Session, err error) {
	op, err := GetCluster(cm)
	if err != nil {
		return
	}
	session, err = op.GetSession(cm, client, db)
	return
}
