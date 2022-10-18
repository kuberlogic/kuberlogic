/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package controllers

import (
	"net/url"
	"time"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	velero "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	appsv1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("KuberlogicService controller", func() {
	const (
		klsName = "test-service"

		defaultReplicas = 1
		defaultDomain   = "example.com"
		//defaultVersion    = "13"
		defaultVolumeSize = "1G"

		interval = time.Millisecond * 250
	)

	timeout := time.Second * 60
	// in a real world backup / restore might take a little long to complete
	if useExistingCluster() {
		timeout = time.Second * 600
	}

	var limits = v1.ResourceList{
		// CPU 250m required minimum for zalando/posgtresql
		// Memory 250Mi required minimum for zalando/posgtresql
		v1.ResourceCPU:     resource.MustParse("250m"),
		v1.ResourceMemory:  resource.MustParse("256Mi"),
		v1.ResourceStorage: resource.MustParse(defaultVolumeSize),
	}

	Context("When updating KuberLogicService", func() {
		kls := &v1alpha1.KuberLogicService{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: klsName,
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

			Expect(k8sClient.Create(ctx, kls)).Should(Succeed())

			By("By checking a new KuberLogicService")
			createdKls := &v1alpha1.KuberLogicService{}

			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(kls), createdKls)
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

			By("Checking scheduled backup cronjob")
			cj := &v12.CronJob{}
			cj.SetName(kls.GetName())
			cj.SetNamespace(kuberlogicNamespace)
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(cj), cj)
			}, timeout, interval).Should(Not(HaveOccurred()))
			Expect(cj.Spec.Schedule).Should(Equal(kls.Spec.BackupSchedule))

			By("Checking file configs")
			cm := &v1.ConfigMap{}
			cm.SetName(kls.GetName())
			cm.SetNamespace(kls.GetName())
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(cm), cm)
			}, timeout, interval).Should(Succeed())

			Expect(k8sClient.Delete(ctx, kls)).Should(Succeed())
		})
	})

	When("testing backup/restore operations with KuberlogicService", func() {
		kls := &v1alpha1.KuberLogicService{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: klsName,
			},
			Spec: v1alpha1.KuberLogicServiceSpec{
				Type:     "docker-compose",
				Replicas: defaultReplicas,
				Limits:   limits,
			},
		}

		It("must be successful", func() {
			By("creating kls")
			Expect(k8sClient.Create(ctx, kls)).Should(Succeed())

			By("waiting until namespace is created")
			Eventually(func() error {
				if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(kls), kls); err != nil {
					return err
				}
				ns := &v1.Namespace{}
				ns.SetName(kls.Status.Namespace)
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(ns), ns)
			}, timeout, interval).Should(Succeed())

			if !useExistingCluster() {
				By("Simulating kls readiness")
				kls.MarkReady("fake")
				kls.Status.Namespace = kls.GetName()
				Expect(k8sClient.Status().Update(ctx, kls)).Should(Succeed())
			}

			By("waiting until kls ready")
			Eventually(func() error {
				if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(kls), kls); err != nil {
					return err
				}
				if r, f, _ := kls.IsReady(); !r {
					return errors.New("service is not ready, status: " + f)
				}
				return nil
			}, timeout, interval).Should(Succeed())

			By("creating backup request")
			klb := &v1alpha1.KuberlogicServiceBackup{}
			klb.SetName(kls.GetName())
			klb.Spec.KuberlogicServiceName = kls.GetName()

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
				By("faking backup pod readiness")
				Eventually(func() error {
					p := &v1.Pod{}
					p.SetName("kl-backup-pod")
					p.SetNamespace(kls.Status.Namespace)

					if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(p), p); err != nil {
						return err
					}
					p.Status.Phase = v1.PodRunning
					return k8sClient.Status().Update(ctx, p)
				}, timeout, interval).Should(Succeed())

				By("Simulating successful backup")
				Eventually(func() error {
					vb := &velero.Backup{}
					vb.SetName(klb.GetName())
					vb.SetNamespace("velero")

					if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(vb), vb); err != nil {
						return err
					}
					vb.Status.Phase = velero.BackupPhaseCompleted
					return k8sClient.Update(ctx, vb)
				}, timeout, interval).Should(Succeed())
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

			By("creating restore request")
			klr := &v1alpha1.KuberlogicServiceRestore{}
			klr.SetName(kls.GetName())
			klr.Spec.KuberlogicServiceBackup = klb.GetName()
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
			}

			By("klr must be successful")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(klr), klr); err != nil {
					return false
				}
				return klr.IsSuccessful()
			}, timeout, interval).Should(BeTrue())

			By("checking klr owner reference (must be owned by klb)")
			Expect(klr.GetOwnerReferences()[0].UID).Should(Equal(klb.GetUID()))

			By("kls restore running status must be false")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(kls), kls); err != nil {
					return false
				}
				restoring, restoreName := kls.RestoreRunning()
				return restoring && restoreName == klr.GetName() && kls.Status.Phase != "Restoring"
			}, timeout, interval).Should(BeFalse())

			By("deleting klb")
			Expect(k8sClient.Delete(ctx, klb)).Should(Succeed())

			if !useExistingCluster() {
				By("Simulating velero deletebackuprequest handling")
				Eventually(func() error {
					dbr := &velero.DeleteBackupRequest{}
					dbr.SetName(klb.GetName())
					dbr.SetNamespace("velero")
					if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(dbr), dbr); err != nil {
						return err
					}
					return k8sClient.Delete(ctx, dbr)
				}, timeout, interval).Should(Succeed())
			}

			if !useExistingCluster() { // FIXME: need to figure out why this is failing on existing cluster on Github Action
				By("checking that klb does not exist")
				Eventually(func() bool {
					return k8serrors.IsNotFound(k8sClient.Get(ctx, client.ObjectKeyFromObject(klb), klb))
				}, timeout, interval).Should(BeTrue())

				Expect(k8sClient.Delete(ctx, kls)).Should(Succeed())
			}
		})
	})

	When("updating application credentials", func() {
		kls := &v1alpha1.KuberLogicService{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: klsName,
			},
			Spec: v1alpha1.KuberLogicServiceSpec{
				Type:     "docker-compose",
				Replicas: defaultReplicas,
				Limits:   limits,
			},
		}

		It("must be successful", func() {
			By("creating kls")
			Expect(k8sClient.Create(ctx, kls)).Should(Succeed())

			By("waiting until namespace is created")
			Eventually(func() error {
				if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(kls), kls); err != nil {
					return err
				}
				ns := &v1.Namespace{}
				ns.SetName(kls.Status.Namespace)
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(ns), ns)
			}).Should(Succeed())

			if !useExistingCluster() {
				By("Simulating kls readiness")
				kls.MarkReady("fake")
				kls.Status.Namespace = kls.GetName()
				Expect(k8sClient.Status().Update(ctx, kls)).Should(Succeed())
			}

			By("creating update-credentials request")
			credUpdateRequest := &v1.Secret{}
			credUpdateRequest.SetName(v1alpha1.CredsUpdateSecretName)
			credUpdateRequest.SetNamespace(kls.GetName())
			credUpdateRequest.StringData = map[string]string{
				"token": "demo",
			}
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
				return k8serrors.IsNotFound(k8sClient.Get(ctx, client.ObjectKeyFromObject(credUpdateRequest), credUpdateRequest))
			}, timeout, interval).Should(BeTrue())

			Expect(k8sClient.Delete(ctx, kls)).Should(Succeed())
		})
	})
	When("move kls to archive status", func() {
		By("creating kls")
		kls := &v1alpha1.KuberLogicService{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: klsName,
			},
			Spec: v1alpha1.KuberLogicServiceSpec{
				Type:     "docker-compose",
				Replicas: defaultReplicas,
				Limits:   limits,
			},
		}
		It("must be successful", func() {
			Expect(k8sClient.Create(ctx, kls)).Should(Succeed())

			By("waiting until kls is ready")
			Eventually(func() error {
				if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(kls), kls); err != nil {
					return err
				}
				if r, f, _ := kls.IsReady(); !r {
					return errors.New("service is not ready, status: " + f)
				}
				return nil
			}, timeout, interval).Should(Succeed())

			By("setting kls to archive")
			kls.Spec.Archived = true
			Expect(k8sClient.Update(ctx, kls)).Should(Succeed())

			By("waiting until kls is archive state")
			Eventually(func() error {
				if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(kls), kls); err != nil {
					return err
				}
				if !kls.Archived() {
					return errors.New("service is not in Archive state")
				}
				return nil
			}, timeout, interval).Should(Succeed())

			By("waiting the namespace is deleted")
			Eventually(func() error {
				ns := &v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: kls.GetName(),
					},
				}
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(ns), ns)
				if !k8serrors.IsNotFound(err) {
					return err
				} else if err == nil {
					return errors.New("namespace is not deleted")
				}
				return nil
			}, timeout, interval).Should(Succeed())

			By("cleanup kls")
			Expect(k8sClient.Delete(ctx, kls)).Should(Succeed())
		})
	})
})
