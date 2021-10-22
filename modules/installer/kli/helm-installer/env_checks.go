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
	"github.com/Masterminds/semver/v3"
	logger "github.com/kuberlogic/kuberlogic/modules/installer/log"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

var (
	// compatible kubernetes version in semver form
	kubeCompatVersions = []string{"1.20.x-0", "1.21.x-0"}
)

// checkKubernetesVersion checks if Kubernetes cluster is compatible with what is found
func checkKubernetesVersion(clientset kubernetes.Interface, log logger.Logger) error {
	kubeVersion, err := getKubernetesVersion(clientset, log)
	if err != nil {
		return errors.Wrap(err, "error finding cluster version")
	}

	if clusterVersionAllowed(kubeVersion) {
		return nil
	}
	log.Infof("Compatible Kubernetes versions: %v", kubeCompatVersions)
	return errors.New("cluster version incompatible")
}

// getKubernetesVersion gets Kubernetes version and transforms it into *semver.Version for comparison
func getKubernetesVersion(clientset kubernetes.Interface, log logger.Logger) (*semver.Version, error) {
	log.Infof("Checking Kubernetes cluster version")
	info, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, err
	}
	log.Infof("Kubernetes version %s is found", info.String())
	clusterVersion, err := semver.NewVersion(info.String())
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling Kubernetes version")
	}
	return clusterVersion, nil
}

// clusterVersionAllowed checks if a cluster version matches supported version
func clusterVersionAllowed(cluster *semver.Version) bool {
	for _, v := range kubeCompatVersions {
		c, _ := semver.NewConstraint(v)
		if c.Check(cluster) {
			return true
		}
	}
	return false
}

// checkDefaultStorageClass checks if a Kubernetes cluster has a default StorageClass
func checkDefaultStorageClass(clientset kubernetes.Interface, log logger.Logger) error {
	log.Infof("Checking default storage class")
	storageClasses, err := getStorageClasses(clientset)
	if err != nil {
		return errors.Wrap(err, "error listing storageclasses")
	}
	if defaultStorageClassFound(storageClasses) != nil {
		log.Debugf("Found StorageClasses: %v", storageClasses)
	}
	if sc := defaultStorageClassFound(storageClasses); sc != nil {
		log.Infof("Found default StorageClass %s", sc.Name)
		return nil
	}
	log.Errorf("Default storage class is not found")
	return errors.New("default storage class is not found")
}

// getStorageClasses lists v1.StorageClass in Kubernetes
func getStorageClasses(clientset kubernetes.Interface) ([]storagev1.StorageClass, error) {
	scList, err := clientset.StorageV1().StorageClasses().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return scList.Items, nil
}

// defaultStorageClassFound searches for a default StorageClasses in StorageClasses list
func defaultStorageClassFound(scs []storagev1.StorageClass) *storagev1.StorageClass {
	for _, sc := range scs {
		if anno := sc.GetAnnotations(); anno["storageclass.kubernetes.io/is-default-class"] != "" {
			return &sc
		}
	}
	return nil
}

// checkLoadBalancerServiceType checks if a cluster provisions IngressIP for test LoadBalancer service
func checkLoadBalancerServiceType(clientset kubernetes.Interface, log logger.Logger) error {
	const (
		testServiceName = "kubernetes-test-service"
		testNamespace   = "default"
		waitTimeoutSec  = 30
	)
	log.Infof("Creating test service %s/%s", testServiceName, testNamespace)
	svc, err := createTestLoadBalancer(testServiceName, testNamespace, clientset)
	if err != nil || svc == nil {
		return errors.Wrap(err, "error creating test service")
	}

	// handle service deletion
	defer func() {
		clientset.CoreV1().Services(testNamespace).Delete(context.TODO(), testServiceName, v1.DeleteOptions{})
		log.Debugf("test service %s deleted", testNamespace, testServiceName)
	}()

	log.Infof("Checking LoadBalancer service type")
	for i := 1; i < waitTimeoutSec; i += 1 {
		time.Sleep(time.Second)
		s, err := clientset.CoreV1().Services(testNamespace).Get(context.TODO(), testServiceName, v1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "error getting test service")
		}
		if loadBalancerProvisioned(s) {
			return nil
		}
		continue
	}
	return errors.New("Service with LoadBalancer didn't get ingress address")
}

func createTestLoadBalancer(name, ns string, clientset kubernetes.Interface) (*corev1.Service, error) {
	const testPort = 9999

	svc := &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Type: "LoadBalancer",
			Ports: []corev1.ServicePort{
				{
					Name:     "test",
					Port:     testPort,
					Protocol: "TCP",
				},
			},
		},
	}

	s, err := clientset.CoreV1().Services(ns).Create(context.TODO(), svc, v1.CreateOptions{})
	return s, err
}

func loadBalancerProvisioned(svc *corev1.Service) bool {
	if len(svc.Status.LoadBalancer.Ingress) != 0 {
		return true
	}

	if len(svc.Spec.ExternalIPs) != 0 {
		return true
	}

	if svc.Spec.ExternalName != "" {
		return true
	}

	return false
}
