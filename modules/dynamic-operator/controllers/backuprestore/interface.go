package backuprestore

import (
	"context"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

type Provider interface {
	BackupRequest(context.Context, *kuberlogiccomv1alpha1.KuberlogicServiceBackup) error
	AfterBackup(context.Context, *kuberlogiccomv1alpha1.KuberlogicServiceBackup) error
	SetKuberlogicBackupStatus(context.Context, *kuberlogiccomv1alpha1.KuberlogicServiceBackup) error
	BackupDeleteRequest(context.Context, *kuberlogiccomv1alpha1.KuberlogicServiceBackup) error
	RestoreRequest(context.Context, *kuberlogiccomv1alpha1.KuberlogicServiceBackup, *kuberlogiccomv1alpha1.KuberlogicServiceRestore) error
	AfterRestore(context.Context, *kuberlogiccomv1alpha1.KuberlogicServiceRestore) error
	SetKuberlogicRestoreStatus(context.Context, *kuberlogiccomv1alpha1.KuberlogicServiceRestore) error
}
