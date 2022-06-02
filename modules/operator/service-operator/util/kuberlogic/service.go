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

// GetCluster returns a service operator based on cm.Spec.Type field
func GetCluster(cm *kuberlogicv1.KuberLogicService) (op interfaces.OperatorInterface, err error) {
	op, err = serviceOperator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm, "")
	return
}

// GetClusterPodLabels returns maps of labels for master and replica service pods
// it returns an error if a cluster Operator is not found
func GetClusterPodLabels(cm *kuberlogicv1.KuberLogicService) (master map[string]string, replica map[string]string, err error) {
	op, err := GetCluster(cm)
	if err != nil {
		return
	}
	master, replica = op.GetInternalDetails().GetPodMasterSelector(), op.GetInternalDetails().GetPodReplicaSelector()

	return
}

// GetClusterServices returns master and replica service names.
// An error is returned if a cluster Operator is not found
func GetClusterServices(cm *kuberlogicv1.KuberLogicService) (master string, replica string, err error) {
	op, err := GetCluster(cm)
	if err != nil {
		return
	}
	master = op.GetInternalDetails().GetMasterService()
	replica = op.GetInternalDetails().GetReplicaService()
	return
}

// GetClusterServicePort returns a service port for a service.
// An error is returned if a cluster Operator is not found
func GetClusterServicePort(cm *kuberlogicv1.KuberLogicService) (p int, err error) {
	op, err := GetCluster(cm)
	if err != nil {
		return
	}
	p = op.GetInternalDetails().GetAccessPort()
	return
}

// GetClusterMainContainer returns a name of a main container for a serivce (mysql for Mysql, etc)
// An error is returned if a cluster Operator is not found
func GetClusterMainContainer(cm *kuberlogicv1.KuberLogicService) (c string, err error) {
	op, err := GetCluster(cm)
	if err != nil {
		return
	}
	c = op.GetInternalDetails().GetMainPodContainer()
	return
}

// GetClusterCredentialsInfo returns:
// * username
// * secret field that contains a password for this username
// * secret name that contains a password
// * error if a service operator is not found
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

// GetSession returns a session struct for a service
// An error is returned if a cluster Operator is not found
func GetSession(cm *kuberlogicv1.KuberLogicService, client kubernetes.Interface, db string) (session interfaces.Session, err error) {
	op, err := GetCluster(cm)
	if err != nil {
		return
	}
	session, err = op.GetSession(cm, client, db)
	return
}
