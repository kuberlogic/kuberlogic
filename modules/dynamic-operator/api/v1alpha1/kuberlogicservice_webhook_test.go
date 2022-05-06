/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

		timeout = time.Second * 10
		//duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	var defaultLimits = v1.ResourceList{
		// CPU 250m required minimum for zalando/posgtresql
		// Memory 250Mi required minimum for zalando/posgtresql
		v1.ResourceCPU:    resource.MustParse("250m"),
		v1.ResourceMemory: resource.MustParse("256Mi"),
	}

	Context("When updating KuberLogicService", func() {
		It("Should create KuberLogicService resource", func() {

			By("By creating a new KuberLogicService")

			//ctx := context.Background()
			kls := &KuberLogicService{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kuberlogic.com/v1alpha1",
					Kind:       "KuberLogicService",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      klsName,
					Namespace: klsNamespace,
				},
				Spec: KuberLogicServiceSpec{
					Type:       "postgresql",
					Replicas:   defaultReplicas,
					VolumeSize: defaultVolumeSize,
					Version:    defaultVersion,
				},
			}

			Expect(k8sClient.Create(ctx, kls)).Should(Succeed())

			By("By checking a new KuberLogicService")
			lookupKlsKey := types.NamespacedName{Name: klsName, Namespace: klsNamespace}
			createdKls := &KuberLogicService{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, lookupKlsKey, createdKls)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			log.Info("resources", "res", createdKls.Spec.Limits)
			Expect(createdKls.Spec.Limits["cpu"]).Should(Equal(defaultLimits["cpu"]))
			Expect(createdKls.Spec.Limits["memory"]).Should(Equal(defaultLimits["memory"]))

			//advanced, _ := json.Marshal(map[string]interface{}{
			//	"resources": defaultResources,
			//})

			// check the defaults is added to configuration
			//Expect(createdKls.Spec.Advanced.Raw).Should(Equal(advanced))
		})
	})
})