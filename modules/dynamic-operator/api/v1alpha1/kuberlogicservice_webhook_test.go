/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var _ = Describe("KuberlogicService controller", func() {
	const (
		klsName      = "test-service"
		klsNamespace = "default"

		defaultReplicas = 1
		//defaultVersion  = "13"

		timeout = time.Second * 10
		//duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	var defaultLimits = v1.ResourceList{
		// CPU 250m required minimum for zalando/posgtresql
		// Memory 250Mi required minimum for zalando/posgtresql
		v1.ResourceCPU:     resource.MustParse("255m"),
		v1.ResourceMemory:  resource.MustParse("356Mi"),
		v1.ResourceStorage: resource.MustParse("10Gi"),
	}

	Context("When updating KuberLogicService", func() {
		//ctx := context.Background()
		kls := &KuberLogicService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      klsName,
				Namespace: klsNamespace,
			},
			Spec: KuberLogicServiceSpec{
				Type:     "docker-compose",
				Replicas: defaultReplicas,
				Limits:   defaultLimits,
			},
		}

		defaultResourceKls := &KuberLogicService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      klsName + "-default-resources",
				Namespace: klsNamespace,
			},
			Spec: KuberLogicServiceSpec{
				Type:     "docker-compose",
				Replicas: defaultReplicas,
			},
		}

		It("Should create KuberLogicService resource", func() {
			By("By creating a new KuberLogicService")

			Expect(k8sClient.Create(ctx, kls)).Should(Succeed())

			By("By checking a new KuberLogicService")
			lookupKlsKey := types.NamespacedName{Name: klsName, Namespace: klsNamespace}
			createdKls := &KuberLogicService{}

			Eventually(func() error {
				return k8sClient.Get(ctx, lookupKlsKey, createdKls)
			}, timeout, interval).Should(Not(HaveOccurred()))

			log.Info("resources", "res", createdKls.Spec.Limits)
			Expect(createdKls.Spec.Limits["cpu"]).Should(Equal(defaultLimits["cpu"]))
			Expect(createdKls.Spec.Limits["memory"]).Should(Equal(defaultLimits["memory"]))
			Expect(createdKls.Spec.Limits["storage"]).Should(Equal(defaultLimits["storage"]))

			By("By creating a new KuberLogicService with default limits")

			Expect(k8sClient.Create(ctx, defaultResourceKls)).Should(Succeed())

			By("By checking default plugin resources")
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(defaultResourceKls), createdKls)
			}, timeout, interval).Should(Not(HaveOccurred()))

			log.Info("resources", "res", createdKls.Spec.Limits)
			Expect(createdKls.Spec.Limits["storage"]).Should(Equal(resource.MustParse("1Gi")))

			By("Volume downsize is not supported")
			defaultResourceKls.Spec.Limits["storage"] = resource.MustParse("1Mi")
			Expect(k8sClient.Update(ctx, defaultResourceKls).Error()).Should(ContainSubstring("volume downsize forbidden"))
		})
		It("Should remove KuberLogicService resource", func() {
			By("Removing KuberLogicService resource")

			for _, item := range []*KuberLogicService{kls, defaultResourceKls} {
				Expect(k8sClient.Delete(ctx, item)).Should(Succeed())

				By("By checking a new KuberLogicService")
				removedKls := &KuberLogicService{}
				Eventually(func() bool {
					return errors.IsNotFound(k8sClient.Get(ctx, client.ObjectKeyFromObject(item), removedKls))
				}, timeout, interval).Should(BeTrue())
			}
		})
	})
})
