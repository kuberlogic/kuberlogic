package backuprestore

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	"github.com/pkg/errors"
	velero "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

const (
	veleroNamespace = "velero"
	veleroStorage   = "default"

	// ResticBackupPodName is used by service-paused webhook as well
	ResticBackupPodName           = "kl-backup-pod"
	resticBackupVolumesAnnotation = "backup.velero.io/backup-volumes"
)

var (
	maxWaitTimeout = 60

	errVeleroBackupIsNotSuccessful = errors.New("velero backup is not successful")

	backupStorageLocationMaxCheckTTL             = 15.0
	errVeleroBackupStorageLocationIsNotAvailable = fmt.Errorf("velero backup storage location is unavailable or checked more than %f minutes ago", backupStorageLocationMaxCheckTTL)

	errBackupPodNotReady = errors.New("backup pod did not become ready")

	errNamespaceExists = errors.New("namespace exists during timeout")
	errPodsNotPaused   = errors.New("found non-pending service pods")
)

type VeleroBackupRestore struct {
	volumeSnapshotsEnabled bool

	kubeClient client.Client
	log        logr.Logger

	kls *kuberlogiccomv1alpha1.KuberLogicService
}

//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservicerestores/finalizers,verbs=update
//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservicerestores,verbs=delete;list
//+kubebuilder:rbac:groups="velero.io",resources=restores;backups;backupstoragelocations;deletebackuprequests;deletebackuprequests/finalizers,verbs=get;create;list;update;watch
//+kubebuilder:rbac:groups="velero.io",resources=deletebackuprequests/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pvc,verbs=list
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;delete;create;list
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;delete

func (v *VeleroBackupRestore) BackupRequest(ctx context.Context, klb *kuberlogiccomv1alpha1.KuberlogicServiceBackup) error {
	log := v.log.WithValues("operation", "BackupRequest")
	log.Info("Started routine")

	veleroBackup := newVeleroBackupObject(klb.GetName(), v.kls)
	veleroBackup.Spec.SnapshotVolumes = &v.volumeSnapshotsEnabled

	veleroBackupStorageLocation := &velero.BackupStorageLocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      veleroBackup.Spec.StorageLocation,
			Namespace: veleroNamespace,
		},
	}
	if err := v.kubeClient.Get(ctx, client.ObjectKeyFromObject(veleroBackupStorageLocation), veleroBackupStorageLocation); err != nil {
		log.Error(err, "failed to get velero backup storage location", "velero backup storage location", veleroBackupStorageLocation)
		return err
	}
	if !isVeleroBackupStorageLocationAvailable(veleroBackupStorageLocation) {
		log.Error(errVeleroBackupStorageLocationIsNotAvailable, "velero backup storage location must be available",
			"velero backup storage location", veleroBackupStorageLocation)
		return errVeleroBackupStorageLocationIsNotAvailable
	}

	// snapshots are disabled. going the restic route
	if !v.volumeSnapshotsEnabled {
		var podsPaused bool
		for waitForPausedPods := maxWaitTimeout; waitForPausedPods > 0; waitForPausedPods -= 1 {
			time.Sleep(1 * time.Second)
			podList := &v1.PodList{}
			log.Info("found pods in namespace", "count", len(podList.Items), "pods", podList.Items)
			if err := v.kubeClient.List(ctx, podList, &client.ListOptions{Namespace: v.kls.Status.Namespace}); err != nil {
				log.Error(err, "failed to list service pods")
				continue
			}

			podsPaused = true
			for _, p := range podList.Items {
				if p.GetName() == ResticBackupPodName {
					continue
				}
				if p.Status.Phase != v1.PodPending {
					podsPaused = false
					err := v.kubeClient.Delete(ctx, &p)
					v.log.Info("tried to delete service pod", "name", p.GetName(), "error", err)
				}
			}
			if podsPaused {
				log.Info("all service pods are paused")
				break
			}
		}
		if !podsPaused {
			log.Error(errPodsNotPaused, "all service pods must be paused")
			return errPodsNotPaused
		}

		// prepare a backup helper pod and mark all volumes that need to be backed up
		backupPod := resticBackupPod(v.kls)

		pvcList := &v1.PersistentVolumeClaimList{}
		if err := v.kubeClient.List(ctx, pvcList, &client.ListOptions{Namespace: v.kls.Status.Namespace}); err != nil {
			log.Error(err, "failed to list PVCs")
			return err
		}
		for _, pvc := range pvcList.Items {
			backupPod.Spec.Volumes = append(backupPod.Spec.Volumes, v1.Volume{
				Name: pvc.GetName(),
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvc.GetName(),
						ReadOnly:  true,
					},
				},
			})
			backupPod.Spec.Containers[0].VolumeMounts = append(backupPod.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
				Name:      pvc.GetName(),
				ReadOnly:  true,
				MountPath: "/" + pvc.GetName(),
			})
			backupPod.Annotations[resticBackupVolumesAnnotation] = backupPod.Annotations[resticBackupVolumesAnnotation] + pvc.GetName() + ","
		}

		err := controllerruntime.SetControllerReference(klb, backupPod, v.kubeClient.Scheme())
		if err != nil {
			return err
		}
		if err := v.kubeClient.Create(ctx, backupPod); err != nil && !k8serrors.IsAlreadyExists(err) {
			log.Error(err, "failed to create backup helper pod", "pod", backupPod)
			return err
		}

		// wait until pod is ready
		for waitForBackupPodSec := maxWaitTimeout; waitForBackupPodSec > 0; waitForBackupPodSec -= 1 {
			time.Sleep(1 * time.Second)
			log.Info("waiting for backup pod to be ready")
			if err := v.kubeClient.Get(ctx, client.ObjectKeyFromObject(backupPod), backupPod); err != nil && !k8serrors.IsNotFound(err) {
				// only log error and continue
				log.Error(err, "error getting backup pod state")
				continue
			}
			if backupPod.Status.Phase == v1.PodRunning {
				log.Info("backup pod ready")
				break
			}
		}
		if backupPod.Status.Phase != v1.PodRunning {
			log.Error(errBackupPodNotReady, "timed out waiting for backup pod")
			return errBackupPodNotReady
		}
	}

	_ = controllerruntime.SetControllerReference(klb, veleroBackup, v.kubeClient.Scheme())
	if err := v.kubeClient.Create(ctx, veleroBackup); err != nil && !k8serrors.IsAlreadyExists(err) {
		log.Error(err, "failed to create velero backup request", "object", veleroBackup)
		return err
	}
	return nil
}

func (v *VeleroBackupRestore) AfterBackup(ctx context.Context, klb *kuberlogiccomv1alpha1.KuberlogicServiceBackup) error {
	log := v.log.WithValues("operation", "AfterBackup")
	log.Info("Started routine")

	// snapshots do not require extra cleaning
	if v.volumeSnapshotsEnabled {
		return nil
	}

	backupPod := resticBackupPod(v.kls)
	if err := v.kubeClient.Get(ctx, client.ObjectKeyFromObject(backupPod), backupPod); k8serrors.IsNotFound(err) {
		// pod not found, do nothing
	} else if err != nil {
		log.Error(err, "failed to get backup pod")
		return err
	} else {
		if err := v.kubeClient.DeleteAllOf(ctx, &v1.Pod{}, &client.DeleteAllOfOptions{
			ListOptions: client.ListOptions{
				Namespace: v.kls.Status.Namespace,
			},
		}); err != nil {
			log.Error(err, "failed to delete pods")
			return err
		}
	}

	return nil
}

func (v *VeleroBackupRestore) SetKuberlogicBackupStatus(ctx context.Context, klb *kuberlogiccomv1alpha1.KuberlogicServiceBackup) error {
	log := v.log.WithValues("operation", "SetKuberlogicBackupStatus")
	log.Info("Started routine")

	veleroBackup := newVeleroBackupObject(klb.GetName(), v.kls)
	if err := v.kubeClient.Get(ctx, client.ObjectKeyFromObject(veleroBackup), veleroBackup); err != nil {
		if k8serrors.IsNotFound(err) {
			klb.MarkPending()
			return v.kubeClient.Status().Update(ctx, klb)
		}
		return err
	}

	log.Info("velero backup", "object", veleroBackup)
	if isVeleroBackupSuccessful(veleroBackup) {
		klb.MarkSuccessful()
	} else if failed, reason := isVeleroBackupFailed(veleroBackup); failed {
		klb.MarkFailed(reason)
	} else {
		klb.MarkRequested()
	}

	return v.kubeClient.Status().Update(ctx, klb)
}

func (v *VeleroBackupRestore) BackupDeleteRequest(ctx context.Context, klb *kuberlogiccomv1alpha1.KuberlogicServiceBackup) error {
	log := v.log.WithValues("operation", "DeleteRequest")
	log.Info("Started routine")

	veleroBackup := newVeleroBackupObject(klb.GetName(), v.kls)

	deleteRequest := &velero.DeleteBackupRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      klb.GetName(),
			Namespace: veleroNamespace,
		},
		Spec: velero.DeleteBackupRequestSpec{
			BackupName: veleroBackup.GetName(),
		},
	}
	controllerutil.AddFinalizer(deleteRequest, BackupDeleteFinalizer)
	_ = controllerruntime.SetControllerReference(klb, deleteRequest, v.kubeClient.Scheme())

	if err := v.kubeClient.Get(ctx, client.ObjectKeyFromObject(deleteRequest), deleteRequest); k8serrors.IsNotFound(err) {
		if err := v.kubeClient.Create(ctx, deleteRequest); err != nil {
			log.Error(err, "failed to create velero backup delete request")
			return err
		}
	} else if err != nil {
		log.Error(err, "failed to get velero backup delete request")
		return err
	}

	if !deleteRequest.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("backup delete request has been fulfilled")
		controllerutil.RemoveFinalizer(deleteRequest, BackupDeleteFinalizer)
		if err := v.kubeClient.Update(ctx, deleteRequest); err != nil {
			v.log.Error(err, "failed to remove velero backup delete request finalizer")
			return err
		}
	} else {
		log.Info("backup delete request has not yet been fulfilled, will retry")
		return nil
	}

	log.Info("removing klb delete finalizer")
	controllerutil.RemoveFinalizer(klb, BackupDeleteFinalizer)
	return v.kubeClient.Update(ctx, klb)
}

func (v *VeleroBackupRestore) RestoreRequest(ctx context.Context, klb *kuberlogiccomv1alpha1.KuberlogicServiceBackup, klr *kuberlogiccomv1alpha1.KuberlogicServiceRestore) error {
	log := v.log.WithValues("operation", "RestoreRequest")

	veleroBackup := newVeleroBackupObject(klb.GetName(), v.kls)
	if err := v.kubeClient.Get(ctx, client.ObjectKeyFromObject(veleroBackup), veleroBackup); err != nil {
		log.Error(err, "failed to get velero backup object", "velero backup", veleroBackup)
		return err
	}
	if !isVeleroBackupSuccessful(veleroBackup) {
		log.Error(errVeleroBackupIsNotSuccessful,
			"velero backup must be successful", "velero backup", veleroBackup)
		return errVeleroBackupIsNotSuccessful
	}

	veleroBackupStorageLocation := &velero.BackupStorageLocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      veleroBackup.Spec.StorageLocation,
			Namespace: veleroNamespace,
		},
	}
	if err := v.kubeClient.Get(ctx, client.ObjectKeyFromObject(veleroBackupStorageLocation), veleroBackupStorageLocation); err != nil {
		log.Error(err, "failed to get velero backup storage location", "velero backup storage location", veleroBackupStorageLocation)
		return err
	}
	if !isVeleroBackupStorageLocationAvailable(veleroBackupStorageLocation) {
		log.Error(errVeleroBackupStorageLocationIsNotAvailable, "velero backup storage location must be available",
			"velero backup storage location", veleroBackupStorageLocation)
		return errVeleroBackupStorageLocationIsNotAvailable
	}

	// now delete kls namespace (it will be created with restore)
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: v.kls.Status.Namespace,
		},
	}
	if err := v.kubeClient.Delete(ctx, ns); err != nil && !k8serrors.IsNotFound(err) {
		log.Error(err, "failed to delete kls namespace for restore")
		return err
	}

	nsDeleted := false
	for waitForNsDeletion := maxWaitTimeout; waitForNsDeletion > 0; waitForNsDeletion -= 1 {
		v.log.Info("waiting for namespace to be deleted", "namespace", ns)
		time.Sleep(time.Second * 1)
		if err := v.kubeClient.Get(ctx, client.ObjectKeyFromObject(ns), ns); err != nil {
			if k8serrors.IsNotFound(err) {
				log.Info("namespace is deleted")
				nsDeleted = true
				break
			}
			log.Error(err, "failed to get namespace", "namespace", ns)
		}
	}
	if !nsDeleted {
		return errors.Wrap(errNamespaceExists, fmt.Sprintf("namespace %s was not deleted within timeout", ns))
	}

	// create velero restore object
	veleroRestore := newVeleroRestoreObject(klr)
	_ = controllerruntime.SetControllerReference(klr, veleroRestore, v.kubeClient.Scheme())
	if err := v.kubeClient.Create(ctx, veleroRestore); err != nil && !k8serrors.IsAlreadyExists(err) {
		log.Error(err, "failed to create velero restore", "velero restore", veleroRestore)
		return err
	}

	_ = controllerruntime.SetControllerReference(klb, klr, v.kubeClient.Scheme())
	return v.kubeClient.Update(ctx, klr)
}

func (v *VeleroBackupRestore) AfterRestore(ctx context.Context, klr *kuberlogiccomv1alpha1.KuberlogicServiceRestore) error {
	log := v.log.WithValues("operation", "AfterRestore")
	log.Info("Started routine")

	// delete backup pod if exists
	pod := resticBackupPod(v.kls)
	if err := v.kubeClient.Get(ctx, client.ObjectKeyFromObject(pod), pod); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		log.Error(err, "failed to get backup pod", "backup pod", pod)
		return err
	}
	if err := v.kubeClient.DeleteAllOf(ctx, &v1.Pod{}, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			Namespace: v.kls.Status.Namespace,
		},
	}); err != nil && !k8serrors.IsNotFound(err) {
		log.Error(err, "failed to delete backup pod", "backup pod", pod)
		return err
	}

	return nil
}

func (v *VeleroBackupRestore) SetKuberlogicRestoreStatus(ctx context.Context, klr *kuberlogiccomv1alpha1.KuberlogicServiceRestore) error {
	log := v.log.WithValues("operation", "SetKuberlogicRestoreStatus")
	log.Info("Started routine")

	veleroRestore := newVeleroRestoreObject(klr)
	if err := v.kubeClient.Get(ctx, client.ObjectKeyFromObject(veleroRestore), veleroRestore); err != nil {
		if k8serrors.IsNotFound(err) {
			klr.MarkPending()
			return v.kubeClient.Status().Update(ctx, klr)
		}
		log.Error(err, "failed to get velero restore", "velero restore", veleroRestore)
		return err
	}

	if failed, reason := isVeleroRestoreFailed(veleroRestore); failed {
		klr.MarkFailed(reason)
	} else if isVeleroRestoreSuccessful(veleroRestore) {
		klr.MarkSuccessful()
	} else {
		klr.MarkRequested()
	}

	return v.kubeClient.Status().Update(ctx, klr)
}

func newVeleroRestoreObject(klr *kuberlogiccomv1alpha1.KuberlogicServiceRestore) *velero.Restore {
	return &velero.Restore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      klr.GetName(),
			Namespace: veleroNamespace,
		},
		Spec: velero.RestoreSpec{
			BackupName: klr.Spec.KuberlogicServiceBackup,
			ExcludedResources: []string{
				"nodes",
				"events",
				"events.events.k8s.io",
				"backups.velero.io",
				"restores.velero.io",
				"resticrepositories.velero.io",
			},
			IncludedNamespaces: []string{"*"},
		},
	}
}

func newVeleroBackupObject(name string, kls *kuberlogiccomv1alpha1.KuberLogicService) *velero.Backup {

	return &velero.Backup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: veleroNamespace,
		},
		Spec: velero.BackupSpec{
			Metadata:                velero.Metadata{},
			IncludedNamespaces:      []string{kls.Status.Namespace},
			Hooks:                   velero.BackupHooks{},
			StorageLocation:         veleroStorage,
			VolumeSnapshotLocations: []string{veleroStorage},
		},
	}
}

func isVeleroBackupSuccessful(v *velero.Backup) bool {
	return v.Status.CompletionTimestamp != nil && v.Status.Progress != nil &&
		v.Status.Progress.ItemsBackedUp == v.Status.Progress.TotalItems &&
		v.Status.Phase == velero.BackupPhaseCompleted
}

func isVeleroBackupFailed(v *velero.Backup) (bool, string) {
	if v.Status.CompletionTimestamp != nil {
		if v.Status.Progress != nil && v.Status.Progress.TotalItems != v.Status.Progress.ItemsBackedUp {
			return true, "some items were not backed up"
		}
		return true, "backup failed"
	}
	return false, ""
}

func isVeleroBackupStorageLocationAvailable(v *velero.BackupStorageLocation) bool {
	return v.Status.Phase == velero.BackupStorageLocationPhaseAvailable &&
		v.Status.LastValidationTime != nil &&
		time.Since(v.Status.LastValidationTime.Time).Minutes() < backupStorageLocationMaxCheckTTL
}

func isVeleroRestoreSuccessful(v *velero.Restore) bool {
	return v.Status.CompletionTimestamp != nil &&
		v.Status.Errors == 0 &&
		v.Status.Progress != nil && v.Status.Progress.ItemsRestored == v.Status.Progress.TotalItems &&
		v.Status.Phase == velero.RestorePhaseCompleted
}

func isVeleroRestoreFailed(v *velero.Restore) (bool, string) {
	return v.Status.CompletionTimestamp != nil &&
		(v.Status.Errors != 0 || v.Status.Phase == velero.RestorePhaseFailed), v.Status.FailureReason
}

func resticBackupPod(kls *kuberlogiccomv1alpha1.KuberLogicService) *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ResticBackupPodName,
			Namespace: kls.Status.Namespace,
			Annotations: map[string]string{
				resticBackupVolumesAnnotation: "",
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "backup-idle",
					Image:           "alpine",
					Command:         []string{"sleep", "3600"},
					ImagePullPolicy: v1.PullIfNotPresent,
				},
			},
		},
	}

	return pod
}
