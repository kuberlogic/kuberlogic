/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package controllers

import (
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	velero "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	postgresv1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	v12 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var _ = Describe("KuberlogicService controller", func() {
	const (
		klsName = "test-service"

		defaultReplicas   = 1
		defaultDomain     = "example.com"
		defaultVersion    = "13"
		defaultVolumeSize = "1G"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	var limits = v1.ResourceList{
		// CPU 250m required minimum for zalando/posgtresql
		// Memory 250Mi required minimum for zalando/posgtresql
		v1.ResourceCPU:    resource.MustParse("250m"),
		v1.ResourceMemory: resource.MustParse("256Mi"),
	}

	Context("When updating KuberLogicService", func() {
		kls := &v1alpha1.KuberLogicService{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "kuberlogic.com/v1alpha1",
				Kind:       "KuberLogicService",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: klsName,
			},
			Spec: v1alpha1.KuberLogicServiceSpec{
				Type:           "postgresql",
				Replicas:       defaultReplicas,
				Domain:         defaultDomain,
				VolumeSize:     defaultVolumeSize,
				Version:        defaultVersion,
				Limits:         limits,
				BackupSchedule: "*/10 * * * *",
				//Advanced:   advanced,
			},
		}

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

			Expect(k8sClient.Create(ctx, kls)).Should(Succeed())

			By("By checking a new KuberLogicService")
			lookupKlsKey := types.NamespacedName{Name: klsName, Namespace: klsName}
			createdKls := &v1alpha1.KuberLogicService{}

			Eventually(func() bool {
				return k8sClient.Get(ctx, lookupKlsKey, createdKls) == nil
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
				return k8sClient.Get(ctx, lookupKey, svc) == nil
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

		It("Scheduled backup job should be created", func() {
			By("Checking scheduled backup cronjob")
			cj := &v12.CronJob{}
			cj.SetName(kls.GetName())
			cj.SetNamespace("default")
			Eventually(func() bool {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(cj), cj) == nil
			}, timeout, interval).Should(BeTrue())
			Expect(cj.Spec.Schedule).Should(Equal(kls.Spec.BackupSchedule))
		})

		When("testing backup/restore", func() {
			klb := &v1alpha1.KuberlogicServiceBackup{}
			klb.SetName(kls.GetName())
			klb.Spec.KuberlogicServiceName = kls.GetName()

			klr := &v1alpha1.KuberlogicServiceRestore{}
			klr.SetName(kls.GetName())
			klr.Spec.KuberlogicServiceBackup = klb.GetName()

			It("must prepare velero env", func() {
				ns := &v1.Namespace{}
				ns.SetName("velero")
				Expect(k8sClient.Create(ctx, ns)).Should(Succeed())
			})

			When("triggering backup", func() {
				It("backup must be successful", func() {
					Expect(k8sClient.Create(ctx, klb)).Should(Succeed())

					By("kls kls backup running status must be true")
					Eventually(func() bool {
						if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(kls), kls); err != nil {
							return false
						}
						backingUp, backup := kls.BackupRunning()
						return backingUp && backup == klb.GetName()
					}, timeout, interval).Should(BeTrue())

					By("Simulating successful backup")
					vb := &velero.Backup{}
					vb.SetName(klb.GetName())
					vb.SetNamespace("velero")
					vb.Status.Phase = velero.BackupPhaseCompleted
					_ = controllerruntime.SetControllerReference(klb, vb, k8sClient.Scheme())

					Expect(k8sClient.Create(ctx, vb)).Should(Succeed())

					By("klb must be successful")
					Eventually(func() bool {
						if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(klb), klb); err != nil {
							return false
						}
						return klb.IsSuccessful()
					}, timeout, interval).Should(BeTrue())

					By("kls backup running status must be false")
					Eventually(func() bool {
						if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(kls), kls); err != nil {
							return false
						}
						backingUp, backup := kls.BackupRunning()
						return backingUp && backup == klb.GetName()
					}, timeout, interval).Should(BeFalse())
				})
			})

			When("triggering restore", func() {
				It("restore should be successful", func() {
					Expect(k8sClient.Create(ctx, klr)).Should(Succeed())

					By("kls restore running status must be true")
					Eventually(func() bool {
						if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(kls), kls); err != nil {
							return false
						}
						restoring, restoreName := kls.RestoreRunning()
						return restoring && restoreName == klr.GetName()
					}, timeout, interval).Should(BeTrue())

					By("simulating successful restore")
					vr := &velero.Restore{}
					vr.SetName(klr.GetName())
					vr.SetNamespace("velero")
					vr.Status.Phase = velero.RestorePhaseCompleted
					_ = controllerruntime.SetControllerReference(klr, vr, k8sClient.Scheme())
					Expect(k8sClient.Create(ctx, vr))

					By("klr must be successful")
					Eventually(func() bool {
						if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(klr), klr); err != nil {
							return false
						}
						return klr.IsSuccessful()
					}, timeout, interval).Should(BeTrue())

					By("kls restore running status must be false")
					Eventually(func() bool {
						if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(kls), kls); err != nil {
							return false
						}
						restoring, restoreName := kls.RestoreRunning()
						return restoring && restoreName == klr.GetName()
					}, timeout, interval).Should(BeFalse())
				})
			})

			When("deleting backup", func() {
				It("should be successful", func() {
					Expect(k8sClient.Delete(ctx, klb)).Should(Succeed())

					By("simulating velero backup delete")
					Eventually(func() error {
						if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(klb), klb); err != nil {
							if errors.IsNotFound(err) {
								return nil
							}
							return err
						}
						klb.Finalizers = []string{}
						return k8sClient.Update(ctx, klb)
					}, timeout, interval).Should(Succeed())

					By("klb and related klr must be deleted")
					Eventually(func() bool {
						if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(klb), klb); !errors.IsNotFound(err) {
							return false
						}
						if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(klr), klr); !errors.IsNotFound(err) {
							return false
						}
						return true
					}, timeout, interval)
				})
			})
		})
	})
})
