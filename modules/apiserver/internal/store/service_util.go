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

package store

import (
	"context"
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/models"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/logging"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/util/k8s"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	util "github.com/kuberlogic/kuberlogic/modules/operator/service-operator/util/kuberlogic"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strconv"
)

// getServiceExternalConnection discovers master/replica services and returns external connection info: host, port, user, etc
func getServiceExternalConnection(c *kubernetes.Clientset, log logging.Logger, kls *kuberlogicv1.KuberLogicService) (*models.ServiceExternalConnection, error) {
	svc := new(models.ServiceExternalConnection)

	masterSvcName, replicaSvcName, err := util.GetClusterServices(kls)
	if err != nil {
		return svc, err
	}
	log.Debugw("services", "master", masterSvcName, "replica", replicaSvcName)

	port, err := util.GetClusterServicePort(kls)
	if err != nil {
		return svc, err
	}

	masterSvc, err := c.CoreV1().Services(kls.Namespace).Get(context.TODO(), masterSvcName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error getting master service")
	}
	replicaSvc, err := c.CoreV1().Services(kls.Namespace).Get(context.TODO(), replicaSvcName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error getting replica service")
	}

	user, password, err := getServiceCredentials(c, log, kls)
	if err != nil {
		return nil, err
	}

	svc.Master = &models.Connection{
		Cert:     "",
		Host:     k8s.GetServiceExternalAddr(masterSvc, log),
		Password: password,
		Port:     int64(port),
		SslMode:  "",
		User:     user,
	}
	svc.Replica = &models.Connection{
		Cert:     "",
		Host:     k8s.GetServiceExternalAddr(replicaSvc, log),
		Password: password,
		Port:     int64(port),
		SslMode:  "",
		User:     user,
	}
	return svc, nil
}

func getServiceInternalConnection(c *kubernetes.Clientset, log logging.Logger, kls *kuberlogicv1.KuberLogicService) (*models.ServiceInternalConnection, error) {
	svc := new(models.ServiceInternalConnection)

	masterSvc, replicaSvc, err := util.GetClusterServices(kls)
	if err != nil {
		return svc, err
	}

	port, err := util.GetClusterServicePort(kls)
	if err != nil {
		return svc, err
	}

	user, password, err := getServiceCredentials(c, log, kls)
	if err != nil {
		return nil, err
	}

	svc.Master = &models.Connection{
		Cert:     "",
		Host:     masterSvc + "." + kls.Namespace,
		Password: password,
		Port:     int64(port),
		SslMode:  "",
		User:     user,
	}
	svc.Replica = &models.Connection{
		Cert:     "",
		Host:     replicaSvc + "." + kls.Namespace,
		Password: password,
		Port:     int64(port),
		SslMode:  "",
		User:     user,
	}
	return svc, nil
}

func getServiceCredentials(c *kubernetes.Clientset, log logging.Logger, kls *kuberlogicv1.KuberLogicService) (user, password string, err error) {
	user, passwordField, secretName, err := util.GetClusterCredentialsInfo(kls)
	if err != nil {
		err = fmt.Errorf("Error getting connecion credentials: %s", err.Error())
		return
	}

	log.Debugw("trying to get credentials for username",
		"user", user, "secret", user, "password", passwordField)
	password, err = k8s.GetSecretFieldDecoded(c, log, secretName, kls.Namespace, passwordField)
	if err != nil {
		return user, password, err
	}
	return
}

func getServiceInstances(c *kubernetes.Clientset, log logging.Logger, kls *kuberlogicv1.KuberLogicService, ctx context.Context) ([]*models.ServiceInstance, error) {
	masterPods, replicaPods, err := getServicePods(c, log, kls, ctx)
	if err != nil {
		return nil, err
	}

	var instances []*models.ServiceInstance
	instances = append(instances, podsToServiceInstances(masterPods, "master")...)
	instances = append(instances, podsToServiceInstances(replicaPods, "replica")...)

	return instances, nil
}

func podsToServiceInstances(pods *corev1.PodList, role string) (instances []*models.ServiceInstance) {
	for _, p := range pods.Items {
		instances = append(instances, &models.ServiceInstance{
			Name:   p.Name,
			Role:   role,
			Status: podStatusToServiceInstanceStatus(p),
		})
	}
	return instances
}

func podStatusToServiceInstanceStatus(p corev1.Pod) *models.ServiceInstanceStatus {
	s := &models.ServiceInstanceStatus{}
	switch p.Status.Phase {
	case "Pending":
		s.Status = "Pending"
		break
	case "Running":
		s.Status = "Running"
		break
	default:
		s.Status = "Failed"
	}
	return s
}

func getServicePods(c *kubernetes.Clientset, log logging.Logger, kls *kuberlogicv1.KuberLogicService, ctx context.Context) (masterPods *corev1.PodList, replicaPods *corev1.PodList, err error) {
	masterPodSelector, replicaPodSelector, err := util.GetClusterPodLabels(kls)
	if err != nil {
		return
	}

	podListOpts := metav1.ListOptions{
		LabelSelector: k8s.MapToStrSelector(masterPodSelector),
	}
	masterPods, err = c.CoreV1().Pods(kls.Namespace).List(ctx, podListOpts)
	log.Debugw("master pods details",
		"master pods", &masterPods, "pod list options", podListOpts)
	if err != nil {
		return
	}

	podListOpts.LabelSelector = k8s.MapToStrSelector(replicaPodSelector)
	replicaPods, err = c.CoreV1().Pods(kls.Namespace).List(ctx, podListOpts)
	if err != nil {
		return
	}
	log.Debugw("replica pods details", "replica pods", &replicaPods, "pod list options", podListOpts)
	return
}

func getServiceInstanceLogs(c *kubernetes.Clientset, kls *kuberlogicv1.KuberLogicService, log logging.Logger, ctx context.Context, instance string, lines int64) (logs string, found bool, err error) {
	found = true

	instances, err := getServiceInstances(c, log, kls, ctx)
	if err != nil {
		return logs, false, fmt.Errorf("error getting service instances: %s", err.Error())
	}

	for _, p := range append(instances) {
		if p.Name == instance {
			container, _ := util.GetClusterMainContainer(kls)
			logs, err = k8s.GetPodLogs(c, log, p.Name, container, kls.Namespace, lines)
			if err != nil {
				return
			}
			return
		}
	}
	return logs, false, nil
}

// memoryQuantityAsG returns a string representation of a resource.Quantity
// converted to a G representation
// e.g, 500M = 0.5G, 1000M = 1G
func memoryQuantityAsG(m resource.Quantity) string {
	const resModifier = 1000
	return strconv.FormatFloat(float64(m.Value())/float64(resModifier*resModifier*resModifier), 'f', 2, 64)
}

// cpuQuantityAsCoreShares returns a string representation of a resource.Quantity
// converted to a number of CPU cores assigned
// e.g. 100m = 0.1. 1 = 1
func cpuQuantityAsCoreShares(m resource.Quantity) string {
	return strconv.FormatFloat(float64(m.MilliValue())/float64(1000), 'f', 2, 64)
}

func int64AsPointer(x int64) *int64 {
	return &x
}

func strAsPointer(x string) *string {
	return &x
}
