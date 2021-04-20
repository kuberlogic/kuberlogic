package interfaces

import (
	"github.com/kuberlogic/operator/modules/operator/api/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OperatorInterface interface {
	Name(cm *v1.KuberLogicService) string
	Init(cm *v1.KuberLogicService)
	InitFrom(o runtime.Object)
	Update(cm *v1.KuberLogicService)
	AsRuntimeObject() runtime.Object
	AsMetaObject() metav1.Object
	AsClientObject() client.Object
	IsEqual(cm *v1.KuberLogicService) bool
	IsReady() (bool, string)
	GetDefaults() v1.Defaults

	GetBackupSchedule() BackupSchedule
	GetBackupRestore() BackupRestore
	GetInternalDetails() InternalDetails
	GetSession(cm *v1.KuberLogicService, client *kubernetes.Clientset, db string) (Session, error)
}

type BackupSchedule interface {
	New(backup *v1.KuberLogicBackupSchedule) v1beta1.CronJob
	Init(*v1.KuberLogicBackupSchedule)
	InitFrom(*v1beta1.CronJob)
	IsEqual(cm *v1.KuberLogicBackupSchedule) bool
	Update(cm *v1.KuberLogicBackupSchedule)
	GetCronJob() *v1beta1.CronJob
	IsSuccessful(job *batchv1.Job) bool
	IsRunning(job *batchv1.Job) bool

	SetBackupImage()
	SetBackupEnv(cm *v1.KuberLogicBackupSchedule)
}

type BackupRestore interface {
	New(backup *v1.KuberLogicBackupRestore) batchv1.Job
	Init(*v1.KuberLogicBackupRestore)
	InitFrom(*batchv1.Job)
	GetJob() *batchv1.Job
	IsSuccessful() bool
	IsRunning() bool

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

type Session interface {
	GetDatabase() Database
	GetUser() User
	CreateTable(table string) error
}

type Database interface {
	List() ([]string, error)
	Create(name string) error
	Drop(name string) error
	IsProtected(name string) bool
}

type User interface {
	List() ([]string, error)
	Create(name, password string) error
	Delete(name string) error
	Edit(name, password string) error
	IsProtected(name string) bool
}
