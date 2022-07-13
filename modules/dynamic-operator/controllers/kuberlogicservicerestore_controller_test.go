package controllers

import (
	"context"
	"fmt"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/cfg"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	velero "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sync"
	"time"
)

var _ = Describe("KuberlogicServiceRestore Controller", func() {
	var r *KuberlogicServiceRestoreReconciler
	var ctx context.Context

	var klr *kuberlogiccomv1alpha1.KuberlogicServiceRestore
	var klb *kuberlogiccomv1alpha1.KuberlogicServiceBackup
	var kls *kuberlogiccomv1alpha1.KuberLogicService

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		utilruntime.Must(clientgoscheme.AddToScheme(scheme))
		utilruntime.Must(kuberlogiccomv1alpha1.AddToScheme(scheme))
		utilruntime.Must(velero.AddToScheme(scheme))

		c := fake.NewClientBuilder().WithScheme(scheme).Build()

		cfg := &cfg.Config{}
		cfg.Backups.Enabled = true

		r = &KuberlogicServiceRestoreReconciler{
			Client: c,
			Scheme: scheme,
			Cfg:    cfg,
			mu:     sync.Mutex{},
		}
		ctx = context.TODO()

		klr = new(kuberlogiccomv1alpha1.KuberlogicServiceRestore)
		klr.SetName(fmt.Sprintf("klb-%d", time.Now().Unix()))
		klr.Spec.KuberlogicServiceBackup = fmt.Sprintf("kls-%d", time.Now().Unix())

		klb = new(kuberlogiccomv1alpha1.KuberlogicServiceBackup)
		klb.SetName(klr.Spec.KuberlogicServiceBackup)
		klb.Spec.KuberlogicServiceName = fmt.Sprintf("kls-%d", time.Now().Unix())
		klb.MarkSuccessful()

		kls = new(kuberlogiccomv1alpha1.KuberLogicService)
		kls.SetName(klr.Spec.KuberlogicServiceBackup)
		kls.Spec.Type = "postgresql"

		Expect(r.Create(ctx, kls)).Should(Succeed())
		Expect(r.Create(ctx, klr)).Should(Succeed())
		Expect(r.Create(ctx, klb)).Should(Succeed())
	})

	When("Reconciling klr", func() {
		When("referenced resources do not exist", func() {
			When("kls does not exist", func() {
				It("restore should fail", func() {
					Expect(r.Delete(ctx, kls)).Should(Succeed())
					_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(klr)})
					Expect(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(Equal(fmt.Sprintf("kuberlogicservices.kuberlogic.com \"%s\" not found", kls.GetName())))
				})
			})
			When("klb does not exist", func() {
				It("restore should fail", func() {
					Expect(r.Delete(ctx, klb)).Should(Succeed())
					_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(klr)})
					Expect(err).ShouldNot(BeNil())
				})
			})
		})

		When("kls is backing up", func() {
			It("reconcile should be requeued", func() {
				runningKlb := &kuberlogiccomv1alpha1.KuberlogicServiceBackup{}
				runningKlb.SetName("demo")
				runningKlb.MarkRequested()
				kls.SetBackupStatus(runningKlb)
				Expect(r.Status().Update(ctx, kls)).Should(Succeed())

				res, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(klr)})
				Expect(err).Should(BeNil())
				Expect(res.RequeueAfter).Should(Equal(time.Second * 30))
			})
		})

		When("kls is restoring", func() {
			It("reconcile should be requeued", func() {
				runningKlr := &kuberlogiccomv1alpha1.KuberlogicServiceRestore{}
				runningKlr.SetName("demo")
				runningKlr.MarkRequested()
				kls.SetRestoreStatus(runningKlr)
				Expect(r.Status().Update(ctx, kls)).Should(Succeed())

				res, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(klr)})
				Expect(err).Should(BeNil())
				Expect(res.RequeueAfter).Should(Equal(time.Second * 30))
			})
		})

		When("too many failures happen", func() {
			It("klb should be marked as failed", func() {
				// this will fail because velero is not installed
				for try := 1; try <= 10; try += 1 {
					_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(klr)})
					Expect(err).ShouldNot(BeNil())
				}
				_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(klr)})
				Expect(err).Should(BeNil())
				Expect(r.Get(ctx, client.ObjectKeyFromObject(klr), klr))
				Expect(klr.IsFailed()).Should(BeTrue())
			})
		})
	})
})
