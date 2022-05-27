/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package controllers

import (
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	postgresv1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

var _ = Describe("KuberlogicService controller", func() {
	const (
		klsName = "test-service"

		defaultReplicas   = 1
		defaultVersion    = "13"
		defaultVolumeSize = "1G"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	var limits = v1.ResourceList{
		// CPU 250m required minimum for zalando/posgtresql
		// Memory 250Mi required minimum for zalando/posgtresql
		v1.ResourceCPU:    resource.MustParse("250m"),
		v1.ResourceMemory: resource.MustParse("256Mi"),
	}

	Context("When updating KuberLogicService", func() {
		It("Should create KuberLogicService resource", func() {

			By("By creating a new KuberLogicService")

			//defaultResourcesBytes, _ := json.Marshal(defaultResources)
			//rawAdvanced := map[string]interface{}{
			//	"resources": defaultResources,
			//}
			//
			//advancedBytes, _ := json.Marshal(rawAdvanced)
			//advanced := apiextensionsv1.JSON{
			//	Raw: advancedBytes,
			//}

			//ctx := context.Background()
			kls := &v1alpha1.KuberLogicService{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kuberlogic.com/v1alpha1",
					Kind:       "KuberLogicService",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: klsName,
				},
				Spec: v1alpha1.KuberLogicServiceSpec{
					Type:       "postgresql",
					Replicas:   defaultReplicas,
					VolumeSize: defaultVolumeSize,
					Version:    defaultVersion,
					Limits:     limits,
					//Advanced:   advanced,
				},
			}

			Expect(k8sClient.Create(ctx, kls)).Should(Succeed())

			By("By checking a new KuberLogicService")
			lookupKlsKey := types.NamespacedName{Name: klsName, Namespace: klsName}
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
			lookupKey := types.NamespacedName{Name: klsName, Namespace: klsName}

			svc := &unstructured.Unstructured{}
			svc.SetGroupVersionKind(
				postgresv1.SchemeGroupVersion.WithKind(postgresv1.PostgresCRDResourceKind),
			)
			svc.SetName(kls.Name)
			svc.SetNamespace(kls.Name)

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
			//Expect(pgSpec["resources"]).Should(Equal(resources))
			Expect(pgSpec["resources"]).Should(Equal(map[string]interface{}{
				"limits": map[string]interface{}{
					"cpu":    "250m",
					"memory": "256Mi",
				},
				"requests": map[string]interface{}{
					"memory": "128Mi",
					"cpu":    "100m",
				},
			}))
		})
	})
})
