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
			Status: kuberlogiccomv1alpha1.KuberlogicServiceBackupStatus{
				BackupReference: "test",
			},
		}
		klr = &kuberlogiccomv1alpha1.KuberlogicServiceRestore{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Spec: kuberlogiccomv1alpha1.KuberlogicServiceRestoreSpec{
				KuberlogicServiceBackup: klb.GetName(),
			},
			Status: kuberlogiccomv1alpha1.KuberlogicServiceRestoreStatus{
				RestoreReference: "test",
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

				Expect(errors.Is(backupRestore.BackupRequest(ctx, klb), errServicePodsFound)).Should(BeTrue())
				By("Service pod should not exist")
				Expect(errors2.IsNotFound(fakeClient.Get(ctx, client.ObjectKeyFromObject(svcPod), svcPod))).Should(BeTrue())

				Expect(backupRestore.BackupRequest(ctx, klb)).Should(Succeed())
				By("Backup pod is ready")
				backupPod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      ResticBackupPodName,
						Namespace: ns.GetName(),
					},
				}
				Expect(fakeClient.Get(ctx, client.ObjectKeyFromObject(backupPod), backupPod)).Should(Succeed())
				backupPod.Status.Phase = corev1.PodRunning
				Expect(fakeClient.Update(ctx, backupPod))
				Expect(backupRestore.BackupRequest(ctx, klb)).Should(Succeed())

				By("Checking backup pod volume mounts")
				_ = fakeClient.Get(ctx, client.ObjectKeyFromObject(backupPod), backupPod)
				Expect(backupPod.Spec.Volumes[0].PersistentVolumeClaim.ClaimName).Should(Equal(backupPVC.GetName()))

				By("Checking velero backup existence")
				veleroBackup = &velero.Backup{}
				veleroBackup.SetName(klb.Status.BackupReference)
				veleroBackup.SetNamespace(veleroNamespace)
				Expect(fakeClient.Get(ctx, client.ObjectKeyFromObject(veleroBackup), veleroBackup)).Should(Succeed())

				By("Updating velero backup status")
				veleroBackup.Status.Phase = velero.BackupPhaseInProgress
				Expect(fakeClient.Update(ctx, veleroBackup)).Should(Succeed())

				By("klb status must be requested")
				Expect(backupRestore.SetKuberlogicBackupStatus(ctx, klb)).Should(Succeed())
				Expect(klb.IsRequested()).Should(Equal(true))
			})
		})
	})
	When("Backup is finished", func() {
		It("Should clean up successfully", func() {
			backupPod := resticBackupPod(kls)
			Expect(fakeClient.Create(ctx, backupPod)).Should(Succeed())

			By("Deleting pods if backup pod exists")
			Expect(backupRestore.AfterBackup(ctx, klb)).Should(Succeed())
			Expect(errors2.IsNotFound(fakeClient.Get(ctx, client.ObjectKeyFromObject(backupPod), backupPod))).Should(BeTrue())
		})
	})
	When("Backup delete requested", func() {
		It("Should be successful", func() {
			Expect(fakeClient.Create(ctx, veleroBackup)).Should(Succeed())

			Expect(backupRestore.BackupDeleteRequest(ctx, klb)).Should(BeNil())

			By("Checking delete backup request finalizer")
			bdr := &velero.DeleteBackupRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      klb.GetName(),
					Namespace: "velero",
				},
			}
			Expect(fakeClient.Get(ctx, client.ObjectKeyFromObject(bdr), bdr)).Should(Succeed())
			Expect(controllerutil.ContainsFinalizer(bdr, BackupDeleteFinalizer)).Should(Equal(true))

			By("backup delete backup request is being deleted")
			bdr.DeletionTimestamp = &metav1.Time{Time: time.Now()}
			Expect(fakeClient.Update(ctx, bdr)).Should(Succeed())
			Expect(backupRestore.BackupDeleteRequest(ctx, klb)).Should(Succeed())

			By("backup delete request should not have finalizer now")
			bdr = &velero.DeleteBackupRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      klb.GetName(),
					Namespace: "velero",
				},
			}
			Expect(errors2.IsNotFound(fakeClient.Get(ctx, client.ObjectKeyFromObject(bdr), bdr))).Should(BeTrue())
		})
	})
	When("Restore requested", func() {
		When("Velero backup is not successful", func() {
			It("Should fail with error", func() {
				veleroBackup.Status.Phase = velero.BackupPhaseFailed
				Expect(fakeClient.Create(ctx, veleroBackup)).Should(Succeed())
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
				Expect(fakeClient.Create(ctx, veleroBackup)).Should(Succeed())

				// first run will only submit a namespace delete request and return with nil
				Expect(backupRestore.RestoreRequest(ctx, klb, klr)).Should(Succeed())
				Expect(backupRestore.RestoreRequest(ctx, klb, klr)).Should(Succeed())

				By("Checking if namespace is deleted")
				err := fakeClient.Get(ctx, client.ObjectKeyFromObject(ns), ns)
				Expect(errors2.IsNotFound(err)).Should(Equal(true))

				By("Checking if velero restore is created")
				Expect(backupRestore.RestoreRequest(ctx, klb, klr)).Should(Succeed())
				err = fakeClient.Get(ctx, client.ObjectKeyFromObject(veleroRestore), veleroRestore)
				Expect(err).Should(BeNil())

				By("Updating velero restore status")
				veleroRestore.Status.Phase = velero.RestorePhaseInProgress
				Expect(fakeClient.Update(ctx, veleroRestore)).Should(Succeed())

				By("Checking if klr status is requested")
				Expect(backupRestore.SetKuberlogicRestoreStatus(ctx, klr)).Should(Succeed())
				Expect(klr.IsRequested()).Should(Equal(true))
			})
		})
	})
	When("Restore is finished", func() {
		It("Should clean up successfully", func() {
			backupPod := resticBackupPod(kls)
			_ = fakeClient.Create(ctx, backupPod)

			By("Deleting pods if backup pod exists")
			Expect(backupRestore.AfterRestore(ctx, klr)).Should(Succeed())
			err := fakeClient.Get(ctx, client.ObjectKeyFromObject(backupPod), backupPod)
			Expect(errors2.IsNotFound(err)).Should(Equal(true))
		})
	})
})
