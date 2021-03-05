package interfaces

import (
	"github.com/kuberlogic/operator/modules/operator/api/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type OperatorInterface interface {
	Name(cm *v1.KuberLogicService) string
	Init(cm *v1.KuberLogicService)
	InitFrom(o runtime.Object)
	Update(cm *v1.KuberLogicService)
	AsRuntimeObject() runtime.Object
	AsMetaObject() metav1.Object
	IsEqual(cm *v1.KuberLogicService) bool
	CurrentStatus() string
	GetDefaults() v1.Defaults

	GetBackupSchedule() BackupSchedule
	GetBackupRestore() BackupRestore
	GetInternalDetails() InternalDetails
}

type BackupSchedule interface {
	New(backup *v1.KuberLogicBackupSchedule) v1beta1.CronJob
	Init(*v1.KuberLogicBackupSchedule)
	InitFrom(*v1beta1.CronJob)
	IsEqual(cm *v1.KuberLogicBackupSchedule) bool
	Update(cm *v1.KuberLogicBackupSchedule)
	GetCronJob() *v1beta1.CronJob
	CurrentStatus(ev batchv1.JobList) string

	SetBackupImage()
	SetBackupEnv(cm *v1.KuberLogicBackupSchedule)
}

type BackupRestore interface {
	New(backup *v1.KuberLogicBackupRestore) batchv1.Job
	Init(*v1.KuberLogicBackupRestore)
	InitFrom(*batchv1.Job)
	GetJob() *batchv1.Job
	CurrentStatus() string

	SetRestoreImage()
	SetRestoreEnv(cm *v1.KuberLogicBackupRestore)
}

type InternalDetails interface {
	GetCredentialsSecret() (*corev1.Secret, error)

	GetPodReplicaSelector() map[string]string
	GetPodMasterSelector() map[string]string

	GetMasterService() string
	GetReplicaService() string
	GetAccessPort() int

	GetDefaultConnectionPassword() (string, string)
	GetMainPodContainer() string
}
