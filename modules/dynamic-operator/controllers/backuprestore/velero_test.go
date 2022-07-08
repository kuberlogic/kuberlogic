package backuprestore

import (
	"context"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	velero "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	corev1 "k8s.io/api/core/v1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

var _ = Describe("Velero BackupRestore provider", func() {
	var ctx context.Context

	var kls *kuberlogiccomv1alpha1.KuberLogicService
	var klb *kuberlogiccomv1alpha1.KuberlogicServiceBackup
	var klr *kuberlogiccomv1alpha1.KuberlogicServiceRestore

	var veleroRestore *velero.Restore
	var veleroBackup *velero.Backup
	var veleroBackupStorageLocation *velero.BackupStorageLocation

	var ns *corev1.Namespace
	var backupPVC *corev1.PersistentVolumeClaim

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(kuberlogiccomv1alpha1.AddToScheme(scheme))
	utilruntime.Must(velero.AddToScheme(scheme))

	b := fake.NewClientBuilder()
	var fakeClient client.Client
	var backupRestore Provider

	// modify wait timeout for tests
	maxWaitTimeout = 10

	BeforeEach(func() {
		kls = &kuberlogiccomv1alpha1.KuberLogicService{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Spec: kuberlogiccomv1alpha1.KuberLogicServiceSpec{},
			Status: kuberlogiccomv1alpha1.KuberLogicServiceStatus{
				Namespace: "test",
			},
		}

		klb = &kuberlogiccomv1alpha1.KuberlogicServiceBackup{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Spec: kuberlogiccomv1alpha1.KuberlogicServiceBackupSpec{
				KuberlogicServiceName: kls.GetName(),
			},
		}
		klr = &kuberlogiccomv1alpha1.KuberlogicServiceRestore{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Spec: kuberlogiccomv1alpha1.KuberlogicServiceRestoreSpec{
				KuberlogicServiceBackup: klb.GetName(),
			},
		}

		veleroRestore = &velero.Restore{
			ObjectMeta: metav1.ObjectMeta{
				Name:      klr.GetName(),
				Namespace: "velero",
			},
		}
		veleroBackup = &velero.Backup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      klb.GetName(),
				Namespace: "velero",
			},
			Spec: velero.BackupSpec{
				StorageLocation: "default",
			},
			Status: velero.BackupStatus{
				CompletionTimestamp: &metav1.Time{Time: time.Now()},
				Progress: &velero.BackupProgress{
					ItemsBackedUp: 1,
					TotalItems:    1,
				},
				Phase: velero.BackupPhaseCompleted,
			},
		}
		veleroBackupStorageLocation = &velero.BackupStorageLocation{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "default",
				Namespace: "velero",
			},
			Spec: velero.BackupStorageLocationSpec{},
			Status: velero.BackupStorageLocationStatus{
				LastValidationTime: &metav1.Time{Time: time.Now()},
				Phase:              velero.BackupStorageLocationPhaseAvailable,
			},
		}

		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: kls.Status.Namespace,
			},
		}
		backupPVC = &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "demo",
				Namespace: ns.GetName(),
			},
		}

		fakeClient = b.WithScheme(scheme).Build()
		ctx = context.TODO()
		backupRestore = NewVeleroBackupRestoreProvider(fakeClient, logger.FromContext(ctx).WithValues("test"), kls, false)

		for _, o := range []client.Object{kls, klb, klr, veleroBackupStorageLocation, ns, backupPVC} {
			_ = fakeClient.Create(ctx, o)
		}
	})

	Describe("Backup requested", func() {
		When("With unavailable backup storage location", func() {
			It("Should fail", func() {
				veleroBackupStorageLocation.Status.Phase = velero.BackupStorageLocationPhaseUnavailable
				_ = fakeClient.Update(ctx, veleroBackupStorageLocation)

				err := backupRestore.BackupRequest(ctx, klb)
				Expect(errors.Is(err, errVeleroBackupStorageLocationIsNotAvailable)).Should(Equal(true))

				err = backupRestore.SetKuberlogicBackupStatus(ctx, klb)
				Expect(err).Should(BeNil())
				Expect(klb.IsPending()).Should(Equal(true))
			})
		})
		When("With available backup storage location", func() {
			It("Should be successful", func() {
				By("Service pods must be deleted")
				svcPod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "service",
						Namespace: ns.GetName(),
					},
				}
				_ = fakeClient.Create(ctx, svcPod)

				err := backupRestore.BackupRequest(ctx, klb)
				By("Service pod should not exist")
				Expect(errors2.IsNotFound(fakeClient.Get(ctx, client.ObjectKeyFromObject(svcPod), svcPod))).Should(Equal(true))

				By("Backup pod must not be ready")
				Expect(errors.Is(err, errBackupPodNotReady)).Should(Equal(true))

				By("Backup pod is ready")
				backupPod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      ResticBackupPodName,
						Namespace: ns.GetName(),
					},
				}
				_ = fakeClient.Get(ctx, client.ObjectKeyFromObject(backupPod), backupPod)
				go func() {
					time.Sleep(5 * time.Second)
					backupPod.Status.Phase = corev1.PodRunning
					_ = fakeClient.Update(ctx, backupPod)
					return
				}()
				err = backupRestore.BackupRequest(ctx, klb)
				Expect(err).Should(BeNil())

				By("Checking backup pod volume mounts")
				_ = fakeClient.Get(ctx, client.ObjectKeyFromObject(backupPod), backupPod)
				Expect(backupPod.Spec.Volumes[0].PersistentVolumeClaim.ClaimName).Should(Equal(backupPVC.GetName()))

				By("Checking velero backup existence")
				veleroBackup = newVeleroBackupObject(klb.GetName(), kls)
				err = fakeClient.Get(ctx, client.ObjectKeyFromObject(veleroBackup), veleroBackup)
				Expect(err).Should(BeNil())

				By("klb status must be requested")
				err = backupRestore.SetKuberlogicBackupStatus(ctx, klb)
				Expect(err).Should(BeNil())
				Expect(klb.IsRequested()).Should(Equal(true))
			})
		})
	})
	When("Backup is finished", func() {
		It("Should clean up successfully", func() {
			backupPod := resticBackupPod(kls)
			_ = fakeClient.Create(ctx, backupPod)

			By("Deleting pods if backup pod exists")
			err := backupRestore.AfterBackup(ctx, klb)
			Expect(err).Should(BeNil())
			err = fakeClient.Get(ctx, client.ObjectKeyFromObject(backupPod), backupPod)
			Expect(errors2.IsNotFound(err)).Should(Equal(true))
		})
	})
	When("Backup delete requested", func() {
		It("Should be successful", func() {
			err := backupRestore.BackupDeleteRequest(ctx, klb)
			Expect(err).Should(BeNil())

			By("Checking delete backup request finalizer")
			bdr := &velero.DeleteBackupRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      klb.GetName(),
					Namespace: "velero",
				},
			}
			_ = fakeClient.Get(ctx, client.ObjectKeyFromObject(bdr), bdr)
			Expect(controllerutil.ContainsFinalizer(bdr, BackupDeleteFinalizer)).Should(Equal(true))

			By("backup delete backup request is being deleted")
			bdr.DeletionTimestamp = &metav1.Time{Time: time.Now()}
			_ = fakeClient.Update(ctx, bdr)
			err = backupRestore.BackupDeleteRequest(ctx, klb)
			Expect(err).Should(BeNil())

			By("backup delete request should not have finalizer now")
			bdr = &velero.DeleteBackupRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      klb.GetName(),
					Namespace: "velero",
				},
			}
			_ = fakeClient.Get(ctx, client.ObjectKeyFromObject(bdr), bdr)
			Expect(controllerutil.ContainsFinalizer(bdr, BackupDeleteFinalizer)).Should(Equal(false))
		})
	})
	When("Restore requested", func() {
		When("Velero backup does not exist", func() {
			It("Should fail with error", func() {
				err := backupRestore.RestoreRequest(ctx, klb, klr)
				Expect(errors2.IsNotFound(err)).Should(Equal(true))

				By("Checking if klr status is pending")
				err = backupRestore.SetKuberlogicRestoreStatus(ctx, klr)
				Expect(err).Should(BeNil())
				Expect(klr.IsFailed() || klr.IsSuccessful() || klr.IsRequested()).Should(Equal(false))
			})
		})
		When("Velero backup is not successful", func() {
			It("Should fail with error", func() {
				veleroBackup.Status.Phase = velero.BackupPhaseFailed
				_ = fakeClient.Create(ctx, veleroBackup)
				err := backupRestore.RestoreRequest(ctx, klb, klr)
				Expect(errors.Is(err, errVeleroBackupIsNotSuccessful)).Should(Equal(true))
			})
		})
		When("backup storage location is not available", func() {
			It("Should fail with error", func() {
				_ = fakeClient.Create(ctx, veleroBackup)
				veleroBackupStorageLocation.Status.Phase = velero.BackupStorageLocationPhaseUnavailable
				_ = fakeClient.Update(ctx, veleroBackupStorageLocation)
				err := backupRestore.RestoreRequest(ctx, klb, klr)
				Expect(errors.Is(err, errVeleroBackupStorageLocationIsNotAvailable)).Should(Equal(true))
			})
		})
		When("backup storage location is available", func() {
			It("Should be successful", func() {
				_ = fakeClient.Create(ctx, veleroBackup)

				err := backupRestore.RestoreRequest(ctx, klb, klr)
				Expect(err).Should(BeNil())

				By("Checking if namespace is deleted")
				err = fakeClient.Get(ctx, client.ObjectKeyFromObject(ns), ns)
				Expect(errors2.IsNotFound(err)).Should(Equal(true))

				By("Checking if velero restore is created")
				err = fakeClient.Get(ctx, client.ObjectKeyFromObject(veleroRestore), veleroRestore)
				Expect(err).Should(BeNil())

				By("Checking if klr is controlled by klb")
				_ = fakeClient.Get(ctx, client.ObjectKeyFromObject(klr), klr)
				Expect(klr.OwnerReferences[0].Kind).Should(Equal("KuberlogicServiceBackup"))

				By("Checking if klr status is requested")
				err = backupRestore.SetKuberlogicRestoreStatus(ctx, klr)
				Expect(err).Should(BeNil())
				Expect(klr.IsRequested()).Should(Equal(true))
			})
		})
	})
	When("Restore is finished", func() {
		It("Should clean up successfully", func() {
			backupPod := resticBackupPod(kls)
			_ = fakeClient.Create(ctx, backupPod)

			By("Deleting pods if backup pod exists")
			err := backupRestore.AfterRestore(ctx, klr)
			Expect(err).Should(BeNil())
			err = fakeClient.Get(ctx, client.ObjectKeyFromObject(backupPod), backupPod)
			Expect(errors2.IsNotFound(err)).Should(Equal(true))
		})
	})
})
