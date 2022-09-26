/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */
package v1alpha1

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("KuberlogicBackupService controller", func() {
	const (
		klbName = "test-backup"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)
	var defaultLimits = v1.ResourceList{
		// CPU 250m required minimum for zalando/posgtresql
		// Memory 250Mi required minimum for zalando/posgtresql
		v1.ResourceCPU:     resource.MustParse("255m"),
		v1.ResourceMemory:  resource.MustParse("356Mi"),
		v1.ResourceStorage: resource.MustParse("10Gi"),
	}
	Context("When creating KuberlogicServiceBackup", func() {
		kls := &KuberLogicService{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: klbName,
			},
			Spec: KuberLogicServiceSpec{
				Type:     "docker-compose",
				Replicas: 1,
				Limits:   defaultLimits,
			},
		}
		klb := &KuberlogicServiceBackup{
			ObjectMeta: metav1.ObjectMeta{
				Name: klbName,
			},
			Spec: KuberlogicServiceBackupSpec{
				KuberlogicServiceName: kls.GetName(),
			},
		}
		klbBroken := &KuberlogicServiceBackup{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: klbName,
			},
			Spec: KuberlogicServiceBackupSpec{
				KuberlogicServiceName: kls.GetName(),
			},
		}
		It("Should create KuberLogicServiceBackup resource", func() {
			By("By creating a new KuberLogicService")
			Expect(testK8sClient.Create(ctx, kls)).Should(Succeed())

			By("By creating a new KuberLogicServiceBackup")
			Expect(testK8sClient.Create(ctx, klb)).Should(Succeed())

			By("By checking a new KuberLogicServiceBackup")
			createdKlb := &KuberlogicServiceBackup{}
			Eventually(
				func() error {
					return testK8sClient.Get(ctx, client.ObjectKeyFromObject(klb), createdKlb)
				},
				timeout,
				interval,
			).Should(Not(HaveOccurred()))
		})
		// This test depends on variable shadowing, so it wont work on real cluster
		if !useExistingCluster() {
			It("Should not create KuberLogicServiceBackup resource", func() {
				By("Creating a new KuberLogicServiceBackup resource")
				backupsEnabled = false
				defer func() { backupsEnabled = true }()
				Expect(testK8sClient.Create(ctx, klbBroken)).Should(Not(Succeed()))
			})
		}
	})
})
