/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package controllers

import (
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	velero "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	appsv1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"net/url"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

var _ = Describe("KuberlogicService controller", func() {
	const (
		klsName = "test-service"

		defaultReplicas = 1
		defaultDomain   = "example.com"
		//defaultVersion    = "13"
		defaultVolumeSize = "1G"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	var limits = v1.ResourceList{
		// CPU 250m required minimum for zalando/posgtresql
		// Memory 250Mi required minimum for zalando/posgtresql
		v1.ResourceCPU:     resource.MustParse("250m"),
		v1.ResourceMemory:  resource.MustParse("256Mi"),
		v1.ResourceStorage: resource.MustParse(defaultVolumeSize),
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
				Type:     "docker-compose",
				Replicas: defaultReplicas,
				Domain:   defaultDomain,
				//Version:        defaultVersion,
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

			Eventually(func() error {
				return k8sClient.Get(ctx, lookupKlsKey, createdKls)
			}, timeout, interval).Should(Not(HaveOccurred()))
			Expect(createdKls.Spec.Type).Should(Equal("docker-compose"))

			By("By checking a new cluster")

			svc := &appsv1.Deployment{}
			svc.SetName(kls.Name)
			svc.SetNamespace(kls.Name)

			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(svc), svc)
			}, time.Second*20, interval).Should(Not(HaveOccurred()))

			pReplicas := int32(defaultReplicas)
			Expect(svc.Spec.Replicas).Should(Equal(&pReplicas))
		})

		It("Scheduled backup job should be created", func() {
			By("Checking scheduled backup cronjob")
			cj := &v12.CronJob{}
			cj.SetName(kls.GetName())
			cj.SetNamespace("default")
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(cj), cj)
			}, timeout, interval).Should(Not(HaveOccurred()))
			Expect(cj.Spec.Schedule).Should(Equal(kls.Spec.BackupSchedule))
		})

		When("testing backup/restore", func() {
			klb := &v1alpha1.KuberlogicServiceBackup{}
			klb.SetName(kls.GetName())
			klb.Spec.KuberlogicServiceName = kls.GetName()

			klr := &v1alpha1.KuberlogicServiceRestore{}
			klr.SetName(kls.GetName())
			klr.Spec.KuberlogicServiceBackup = klb.GetName()

			When("triggering backup", func() {
				It("backup must be successful", func() {
					Expect(k8sClient.Create(ctx, klb)).Should(Succeed())

					By("kls backup running status must be true")
					Eventually(func() bool {
						if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(kls), kls); err != nil {
							return false
						}
						backingUp, backup := kls.BackupRunning()
						return backingUp && backup == klb.GetName() && kls.Status.Phase == "Backing Up"
					}, timeout, interval).Should(BeTrue())

					if !useExistingCluster() {
						By("Simulating successful backup")
						vb := &velero.Backup{}
						vb.SetName(klb.GetName())
						vb.SetNamespace("velero")
						vb.Status.Phase = velero.BackupPhaseCompleted
						_ = controllerruntime.SetControllerReference(klb, vb, k8sClient.Scheme())

						Expect(k8sClient.Create(ctx, vb)).Should(Succeed())
					}

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
						return backingUp && backup == klb.GetName() && kls.Status.Phase != "Backing Up"
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
						return restoring && restoreName == klr.GetName() && kls.Status.Phase == "Restoring"
					}, timeout, interval).Should(BeTrue())

					if !useExistingCluster() {
						By("simulating successful restore")
						vr := &velero.Restore{}
						vr.SetName(klr.GetName())
						vr.SetNamespace("velero")
						vr.Status.Phase = velero.RestorePhaseCompleted
						_ = controllerruntime.SetControllerReference(klr, vr, k8sClient.Scheme())
						Expect(k8sClient.Create(ctx, vr))
						_ = controllerruntime.SetControllerReference(klb, klr, k8sClient.Scheme())
						Expect(k8sClient.Update(ctx, klr))
					}

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
						return restoring && restoreName == klr.GetName() && kls.Status.Phase != "Restoring"
					}, timeout, interval).Should(BeFalse())
				})
			})

			When("deleting backup", func() {
				It("should be successful", func() {
					Expect(k8sClient.Delete(ctx, klb)).Should(Succeed())

					if !useExistingCluster() {
						By("simulating velero backup delete")
						Eventually(func() error {
							if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(klb), klb); err != nil {
								if errors.IsNotFound(err) {
									return nil
								}
								return err
							}
							klb.Finalizers = make([]string, 0)
							return k8sClient.Update(ctx, klb)
						}, timeout, interval).Should(Succeed())
					}

					By("klb must be deleted")
					Eventually(func() bool {
						return errors.IsNotFound(k8sClient.Get(ctx, client.ObjectKeyFromObject(klb), klb)) &&
							klb.GetUID() == klr.OwnerReferences[0].UID // gc doesn't work in envtest, but it does not really matter in this specific test
					}, timeout, interval).Should(BeTrue())
				})
			})
		})

		When("testing update-credentials procedure", func() {
			credUpdateRequest := &v1.Secret{}
			credUpdateRequest.SetName(v1alpha1.CredsUpdateSecretName)
			credUpdateRequest.SetNamespace(kls.GetName())
			credUpdateRequest.StringData = map[string]string{
				"token": "demo",
			}

			It("request should be reconciled", func() {
				By("Creating update-credentials secret")
				Expect(controllerutil.SetControllerReference(kls, credUpdateRequest, k8sClient.Scheme())).Should(BeNil())
				Expect(k8sClient.Create(ctx, credUpdateRequest)).Should(Succeed())

				if !useExistingCluster() {
					// create fake deployment pod
					p := &v1.Pod{}
					p.SetName(kls.GetName())
					p.SetNamespace(kls.GetName())
					p.Spec.Containers = []v1.Container{
						{
							Name:  "test",
							Image: "test",
						},
					}
					p.SetLabels(map[string]string{"docker-compose.service/name": kls.GetName()})
					Expect(k8sClient.Create(ctx, p)).Should(Succeed())

					// fake SPDY executor
					NewRemoteExecutor = func(c *rest.Config, m string, url *url.URL) (remotecommand.Executor, error) {
						return &fakeExecutor{}, nil
					}
				}

				By("Making sure the secret is absent")
				Eventually(func() bool {
					return errors.IsNotFound(k8sClient.Get(ctx, client.ObjectKeyFromObject(credUpdateRequest), credUpdateRequest))
				}, timeout, interval).Should(BeTrue())
			})
		})

		It("Should remove KuberLogicService resource", func() {
			By("Removing KuberLogicService resource")

			Expect(k8sClient.Delete(ctx, kls)).Should(Succeed())

			By("By checking a new KuberLogicService")
			lookupKlsKey := types.NamespacedName{Name: klsName, Namespace: klsName}
			createdKls := &v1alpha1.KuberLogicService{}

			Eventually(func() bool {
				return errors.IsNotFound(k8sClient.Get(ctx, lookupKlsKey, createdKls))
			}, timeout, interval).Should(BeTrue())
		})
	})
})
