package backuprestore

import (
	"github.com/go-logr/logr"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	BackupDeleteFinalizer = "kuberlogic.com/backup-delete-finalizer"
)

func NewVeleroBackupRestoreProvider(c client.Client, l logr.Logger, kls *kuberlogiccomv1alpha1.KuberLogicService, volumeSnapshotsEnabled bool) Provider {
	return &VeleroBackupRestore{
		volumeSnapshotsEnabled: volumeSnapshotsEnabled,
		kubeClient:             c,
		log:                    l,
		kls:                    kls,
	}
}
