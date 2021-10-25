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
	logger "github.com/kuberlogic/kuberlogic/modules/installer/log"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/storage/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
	"strings"
	"testing"
	"time"
)

// TestCheckKubernetesVersion tests checkKubernetesVersion by faking various cluster versions
func TestCheckKubernetesVersion(t *testing.T) {
	client := fake.NewSimpleClientset()
	log := logger.NewLogger(true)

	testVersionAllowed := func(git string) bool {
		v := strings.Split(git, ".")
		client.Discovery().(*fakediscovery.FakeDiscovery).FakedServerVersion = &version.Info{
			Major:      v[0],
			Minor:      v[1],
			GitVersion: git,
		}

		return checkKubernetesVersion(client, log) == nil
	}

	versionsTable := map[string]bool{
		"1.19.0":                    false,
		"1.20.1":                    true,
		"1.20.1-eks":                true,
		"1.21.0":                    true,
		"1.21.0-somelongprerelease": true,
		"1.22.0":                    false,
	}
	for v, expected := range versionsTable {
		if actual := testVersionAllowed(v); actual != expected {
			t.Errorf("error for version %s, expected %t, got %t", v, expected, actual)
		}
	}
}

// TestCheckDefaultStorageClass tests checkDefaultStorageClass by adding storage classes until a default is added
func TestCheckDefaultStorageClass(t *testing.T) {
	client := fake.NewSimpleClientset()
	log := logger.NewLogger(true)

	testDefaultStorageClassFound := func(s *v1.StorageClass) bool {
		// create storage class first
		client.StorageV1().StorageClasses().Create(context.TODO(), s, v12.CreateOptions{})
		return checkDefaultStorageClass(client, log) == nil
	}

	var sc *v1.StorageClass
	var expected bool
	if actual := testDefaultStorageClassFound(nil); actual != expected {
		t.Errorf("error when added StorageClass %v, expected %t, actual %t", sc, expected, actual)
	}

	sc, expected = &v1.StorageClass{
		ObjectMeta: v12.ObjectMeta{
			Name: "not-default",
		},
	}, false
	if actual := testDefaultStorageClassFound(nil); actual != expected {
		t.Errorf("error when added StorageClass %v, expected %t, actual %t", sc, expected, actual)
	}

	sc, expected = &v1.StorageClass{
		ObjectMeta: v12.ObjectMeta{
			Name: "default",
			Annotations: map[string]string{
				"storageclass.kubernetes.io/is-default-class": "true",
			},
		},
	}, false
}

// TestCheckLoadBalancerServiceType tests checkLoadBalancerServiceType
func TestCheckLoadBalancerServiceType(t *testing.T) {
	client := fake.NewSimpleClientset()
	log := logger.NewLogger(true)

	const (
		testServiceName = "kubernetes-test-service"
		testNamespace   = "default"
		testTimeoutSec  = 30
	)

	// faking controller via goroutine
	go func() {
		for i := 0; i < testTimeoutSec; i += 1 {
			s, _ := client.CoreV1().Services(testNamespace).Get(context.TODO(), testServiceName, v12.GetOptions{})
			if s != nil {
				s.Spec.ExternalIPs = []string{"127.0.0.1"}
				client.CoreV1().Services(testNamespace).Update(context.TODO(), s, v12.UpdateOptions{})
				break
			}
			time.Sleep(time.Millisecond * time.Duration(100))
		}
	}()
	if err := checkLoadBalancerServiceType(client, log); err != nil {
		t.Errorf("loadbalancer's ExternalIP not found")
	}

	// faking controller via goroutine
	go func() {
		for i := 0; i < testTimeoutSec; i += 1 {
			s, _ := client.CoreV1().Services(testNamespace).Get(context.TODO(), testServiceName, v12.GetOptions{})
			if s != nil {
				s.Spec.ExternalName = "loadbalancer.example.com"
				client.CoreV1().Services(testNamespace).Update(context.TODO(), s, v12.UpdateOptions{})
				break
			}
			time.Sleep(time.Millisecond * time.Duration(100))
		}
	}()
	if err := checkLoadBalancerServiceType(client, log); err != nil {
		t.Errorf("loadbalancer's ExternalName not found")
	}

	// faking controller via goroutine
	go func() {
		for i := 0; i < testTimeoutSec; i += 1 {
			s, _ := client.CoreV1().Services(testNamespace).Get(context.TODO(), testServiceName, v12.GetOptions{})
			if s != nil {
				s.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{
					{
						IP:       "127.0.0.1",
						Hostname: "test",
					},
				}
				client.CoreV1().Services(testNamespace).Update(context.TODO(), s, v12.UpdateOptions{})
				break
			}
			time.Sleep(time.Millisecond * time.Duration(100))
		}
	}()
	if err := checkLoadBalancerServiceType(client, log); err != nil {
		t.Errorf("loadbalancer's Ingress not found")
	}

	// default case when a service doesn't have
	if err := checkLoadBalancerServiceType(client, log); err == nil {
		t.Errorf("loadbalancer must have been failed")
	}
}
