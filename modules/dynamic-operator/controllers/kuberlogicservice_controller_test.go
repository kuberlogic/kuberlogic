/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"encoding/json"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	postgresv1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

var _ = Describe("KuberlogicService controller", func() {
	const (
		klsName      = "test-service"
		klsNamespace = "default"

		defaultReplicas   = 1
		defaultVersion    = "13"
		defaultVolumeSize = "1G"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	var defaultResources = map[string]interface{}{
		"requests": map[string]interface{}{
			"memory": "50Mi",
			"cpu":    "50m",
		},
		"limits": map[string]interface{}{
			"memory": "128Mi",
			"cpu":    "250m",
		},
	}

	Context("When updating KuberLogicService", func() {
		It("Should create KuberLogicService resource", func() {

			By("By creating a new KuberLogicService")

			//defaultResourcesBytes, _ := json.Marshal(defaultResources)
			rawAdvanced := map[string]interface{}{
				"resources": defaultResources,
			}

			advancedBytes, _ := json.Marshal(rawAdvanced)
			advanced := apiextensionsv1.JSON{
				Raw: advancedBytes,
			}

			//ctx := context.Background()
			kls := &v1alpha1.KuberLogicService{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kuberlogic.com/v1alpha1",
					Kind:       "KuberLogicService",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      klsName,
					Namespace: klsNamespace,
				},
				Spec: v1alpha1.KuberLogicServiceSpec{
					Type:       "postgresql",
					Replicas:   defaultReplicas,
					VolumeSize: defaultVolumeSize,
					Version:    defaultVersion,
					Advanced:   advanced,
				},
			}

			Expect(k8sClient.Create(ctx, kls)).Should(Succeed())

			By("By checking a new KuberLogicService")
			lookupKlsKey := types.NamespacedName{Name: klsName, Namespace: klsNamespace}
			createdKls := &v1alpha1.KuberLogicService{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, lookupKlsKey, createdKls)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdKls.Spec.Type).Should(Equal("postgresql"))

			By("By checking a new cluster")
			lookupKey := types.NamespacedName{Name: klsName, Namespace: klsNamespace}

			svc := &unstructured.Unstructured{}
			svc.SetGroupVersionKind(
				postgresv1.SchemeGroupVersion.WithKind(postgresv1.PostgresCRDResourceKind),
			)
			svc.SetName(kls.Name)
			svc.SetNamespace(kls.Namespace)

			Eventually(func() bool {
				err := k8sClient.Get(ctx, lookupKey, svc)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			pgSpec := svc.UnstructuredContent()["spec"].(map[string]interface{})
			//fmt.Println("------", pgSpec)

			Expect(pgSpec["numberOfInstances"]).Should(Equal(int64(defaultReplicas)))
			postgresqlSection := pgSpec["postgresql"].(map[string]interface{})
			Expect(postgresqlSection["version"]).Should(Equal(defaultVersion))
			Expect(pgSpec["volume"]).Should(Equal(map[string]interface{}{
				"size": defaultVolumeSize,
			}))
			Expect(pgSpec["resources"]).Should(Equal(defaultResources))

		})
	})
})
