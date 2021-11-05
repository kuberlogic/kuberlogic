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

package k8s

import (
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/logging"
	v1 "k8s.io/api/core/v1"
	"testing"
)

func TestGetServiceExternalAddr(t *testing.T) {
	l := logging.WithComponentLogger("tests")
	for expectedHost, input := range map[string]*v1.Service{
		"":          {},
		"127.0.0.1": {Spec: v1.ServiceSpec{ExternalIPs: []string{"127.0.0.1"}}},
		"localhost": {Spec: v1.ServiceSpec{ExternalName: "localhost"}},
		"1.1.1.1": {
			Spec: v1.ServiceSpec{},
			Status: v1.ServiceStatus{
				LoadBalancer: v1.LoadBalancerStatus{
					Ingress: []v1.LoadBalancerIngress{
						{IP: "1.1.1.1", Hostname: "test"},
					}}}},
		"externalhost": {
			Spec: v1.ServiceSpec{},
			Status: v1.ServiceStatus{
				LoadBalancer: v1.LoadBalancerStatus{
					Ingress: []v1.LoadBalancerIngress{
						{Hostname: "externalhost"},
					}}}}} {
		if actual := GetServiceExternalAddr(input, l); actual != expectedHost {
			t.Errorf("actual external host (%s) does not match expected (%s)", actual, expectedHost)
		}
	}
}

func TestMapToStrSelector(t *testing.T) {
	k8sSelector := map[string]string{"app": "test", "env": "testing"}
	k8sSelectorExpected := "app=test,env=testing"
	k8sSelectorActual := MapToStrSelector(k8sSelector)

	if k8sSelectorActual != k8sSelectorExpected {
		t.Errorf("failed, got %s, want %s", k8sSelectorActual, k8sSelectorExpected)
	}
}
