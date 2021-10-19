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

package helm_installer

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver/v3"
	logger "github.com/kuberlogic/kuberlogic/modules/installer/log"
	"github.com/pkg/errors"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

var (
	// compatible kubernetes version in major.minor form
	kubeCompatVersions = []string{"1.20", "1.21"}
)

// checkKubernetesVersion checks if Kubernetes cluster is compatible with what is found
func checkKubernetesVersion(clientset *kubernetes.Clientset, log logger.Logger) error {
	log.Infof("Checking Kubernetes cluster version")
	info, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return errors.Wrap(err, "error checking Kubernetes version")
	}
	log.Infof("Kubernetes version %s is found", info.String())
	clusterVersion, err := semver.NewVersion(info.String())
	if err != nil {
		return errors.Wrap(err, "error marshaling Kubernetes version")
	}
	for _, v := range kubeCompatVersions {
		c, err := semver.NewConstraint(v)
		if err != nil {
			return errors.Wrap(err, "error creating version constraint")
		}
		if c.Check(clusterVersion) {
			return nil
		}
	}
	log.Infof("Compatible Kubernetes versions: %v", kubeCompatVersions)
	return errors.New("cluster version incompatible")
}

// checkDefaultStorageClass checks if a Kubernetes cluster has a default StorageClass
func checkDefaultStorageClass(clientset *kubernetes.Clientset, log logger.Logger) error {
	log.Infof("Checking default storage class")
	scs, err := clientset.StorageV1().StorageClasses().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error listing storage classes: %v", err)
	}
	found := false
	for _, sc := range scs.Items {
		if anno := sc.GetAnnotations(); anno["storageclass.kubernetes.io/is-default-class"] != "" {
			found = true
			log.Infof("Found default storage class: %s", sc.Name)
		}
	}
	if !found {
		log.Errorf("Default storage class is not found")
		return fmt.Errorf("default storage class is not found")
	}
	return nil
}

// checkLoadBalancerServiceType checks if a cluster provisions IngressIP for test LoadBalancer service
func checkLoadBalancerServiceType(clientset *kubernetes.Clientset, log logger.Logger) error {
	const (
		testServiceName = "kubernetes-test-service"
		testNamespace   = "default"
		testPort        = 9999
		waitTimeoutSec  = 30
	)
	log.Infof("Checking LoadBalancer service type")
	svc := &v12.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      testServiceName,
			Namespace: testNamespace,
		},
		Spec: v12.ServiceSpec{
			Type: "LoadBalancer",
			Ports: []v12.ServicePort{
				{
					Name:     "test",
					Port:     testPort,
					Protocol: "TCP",
				},
			},
		},
	}

	if _, err := clientset.CoreV1().Services(testNamespace).Create(context.TODO(), svc, v1.CreateOptions{}); err != nil {
		return errors.Wrap(err, "error creating test service")
	}
	log.Debugf("test service %s/%s created", testNamespace, testServiceName)

	defer func() {
		clientset.CoreV1().Services(testNamespace).Delete(context.TODO(), testServiceName, v1.DeleteOptions{})
		log.Debugf("test service %s deleted", testNamespace, testServiceName)
	}()

	for i := 1; i < waitTimeoutSec; i += 1 {
		time.Sleep(time.Second)
		s, err := clientset.CoreV1().Services(testNamespace).Get(context.TODO(), testServiceName, v1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "error getting test service")
		}
		if len(s.Status.LoadBalancer.Ingress) != 0 {
			log.Infof("Kubernetes cluster supports LoadBalancer Service type")
			return nil
		}
		continue
	}
	return errors.New("Service with LoadBalancer didn't get ingress address")
}
