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
	"context"
	"encoding/json"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

var _ = Describe("KuberlogicService controller", func() {
	const (
		klsName       = "test-service"
		klsNamespace  = "default"
		klstName      = "mysql"
		klstNamespace = "default"

		defaultReplicas   = "1"
		defaultVersion    = "5.7"
		defaultVolumeSize = "1G"

		replicas   = 2
		volumeSize = "2G"
		secretName = "test-secret-mysql-name"

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

	var defaultSpec = map[string]interface{}{
		"secretName": secretName,
	}

	Context("When updating KuberLogicService", func() {
		BeforeEach(func() {
			By("By creating a new KuberLogicType")

			defaultResourcesBytes, _ := json.Marshal(defaultResources)
			defaultSpecBytes, _ := json.Marshal(defaultSpec)
			defaultVolumeSizeBytes, _ := json.Marshal(defaultVolumeSize)
			defaultVersionBytes, _ := json.Marshal(defaultVersion)

			spec := v1alpha1.KuberLogicServiceTypeSpec{
				Type: "mysql",
				Api: v1alpha1.KuberLogicServiceTypeApiRef{
					Group:   "mysql.presslabs.org",
					Version: "v1alpha1",
					Kind:    "MysqlCluster",
				},
				SpecRef: map[string]v1alpha1.KuberlogicServiceTypeParam{
					"replicas": {
						Path: "spec.replicas",
						Type: "float",
						DefaultValue: apiextensionsv1.JSON{
							Raw: []byte(defaultReplicas),
						},
					},
					"version": {
						Path: "spec.mysqlVersion",
						Type: "string",
						DefaultValue: apiextensionsv1.JSON{
							Raw: []byte(defaultVersionBytes),
						},
					},
					"volumeSize": {
						Path: "spec.volumeSpec.persistentVolumeClaim.resources.requests.storage",
						Type: "string",
						DefaultValue: apiextensionsv1.JSON{
							Raw: defaultVolumeSizeBytes,
						},
					},
					"resources": {
						Path: "spec.podSpec.resources",
						Type: "json",
						DefaultValue: apiextensionsv1.JSON{
							Raw: defaultResourcesBytes,
						},
					},
				},
				DefaultSpec: apiextensionsv1.JSON{
					Raw: defaultSpecBytes,
				},
				StatusRef: v1alpha1.KuberlogicServiceTypeStatusRef{
					Conditions: &v1alpha1.KuberLogicServiceTypeConditions{
						Path:           "status.conditions",
						ReadyCondition: "Ready",
						ReadyValue:     "True",
					},
				},
			}

			ctx := context.Background()
			klst := &v1alpha1.KuberLogicServiceType{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kuberlogic.com/v1alpha1",
					Kind:       "KuberLogicServiceType",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      klstName,
					Namespace: klstNamespace,
				},
				Spec: spec,
			}

			Expect(k8sClient.Create(ctx, klst)).Should(Succeed())

			By("By checking a new KuberLogicType")
			lookupKey := types.NamespacedName{Name: klstName, Namespace: klstNamespace}
			createdKlst := &v1alpha1.KuberLogicServiceType{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, lookupKey, createdKlst)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdKlst.Spec).Should(Equal(spec))
		})
		It("Should create KuberLogicService resource", func() {

			By("By creating a new KuberLogicService")

			rawSpec := map[string]interface{}{
				"type":       "mysql",
				"replicas":   replicas,
				"volumeSize": volumeSize,
			}

			specBytes, _ := json.Marshal(rawSpec)
			specKuberlogic := apiextensionsv1.JSON{
				Raw: specBytes,
			}
			//
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
				Spec: specKuberlogic,
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
			Expect(createdKls.Spec).Should(Equal(specKuberlogic))

			var value map[string]interface{}
			err := json.Unmarshal(createdKls.Spec.Raw, &value)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(value["type"]).Should(Equal("mysql"))

			By("By checking a new KuberLogicType")
			lookupKey := types.NamespacedName{Name: klstName, Namespace: klstNamespace}
			createdKlst := &v1alpha1.KuberLogicServiceType{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, lookupKey, createdKlst)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("By checking a new mysql cluster")
			lookupKeyMysql := types.NamespacedName{Name: klsName, Namespace: klsNamespace}

			svc := &unstructured.Unstructured{}
			svc.SetGroupVersionKind(createdKlst.ServiceGVK())
			svc.SetName(kls.Name)
			svc.SetNamespace(kls.Namespace)

			Eventually(func() bool {
				err := k8sClient.Get(ctx, lookupKeyMysql, svc)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			mysqlSpec := svc.UnstructuredContent()["spec"].(map[string]interface{})
			Expect(mysqlSpec["secretName"]).Should(Equal(secretName))
			Expect(mysqlSpec["replicas"]).Should(Equal(int64(replicas)))
			Expect(mysqlSpec["mysqlVersion"]).Should(Equal(defaultVersion))
			Expect(mysqlSpec["volumeSpec"]).Should(Equal(map[string]interface{}{
				"persistentVolumeClaim": map[string]interface{}{
					"resources": map[string]interface{}{
						"requests": map[string]interface{}{
							"storage": volumeSize,
						},
					},
				},
			}))
		})
		AfterEach(func() {
			By("Remove KuberlogicServiceType")

			removedKlst := &v1alpha1.KuberLogicServiceType{
				ObjectMeta: metav1.ObjectMeta{
					Name:      klstName,
					Namespace: klstNamespace,
				},
			}

			Eventually(func() bool {
				err := k8sClient.Delete(ctx, removedKlst)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})
	})
})
