package service_operator

import (
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/service-operator/mysql"
	"github.com/kuberlogic/operator/modules/operator/service-operator/postgresql"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type OperatorInterface interface {
	Name(cm *kuberlogicv1.KuberLogicService) string
	Init(cm *kuberlogicv1.KuberLogicService)
	InitFrom(o runtime.Object)
	Update(cm *kuberlogicv1.KuberLogicService)
	AsRuntimeObject() runtime.Object
	AsMetaObject() metav1.Object
	IsEqual(cm *kuberlogicv1.KuberLogicService) bool
	CurrentStatus() string
	GetDefaults() kuberlogicv1.Defaults

	GetBackupSchedule() BackupSchedule
	GetBackupRestore() BackupRestore
	GetInternalDetails() InternalDetails
}

type BackupSchedule interface {
	New(backup *kuberlogicv1.KuberLogicBackupSchedule) v1beta1.CronJob
	Init(*kuberlogicv1.KuberLogicBackupSchedule)
	InitFrom(*v1beta1.CronJob)
	IsEqual(cm *kuberlogicv1.KuberLogicBackupSchedule) bool
	Update(cm *kuberlogicv1.KuberLogicBackupSchedule)
	GetCronJob() *v1beta1.CronJob
	CurrentStatus(ev batchv1.JobList) string

	SetBackupImage()
	SetBackupEnv(cm *kuberlogicv1.KuberLogicBackupSchedule)
}

type BackupRestore interface {
	New(backup *kuberlogicv1.KuberLogicBackupRestore) batchv1.Job
	Init(*kuberlogicv1.KuberLogicBackupRestore)
	InitFrom(*batchv1.Job)
	GetJob() *batchv1.Job
	CurrentStatus() string

	SetRestoreImage()
	SetRestoreEnv(cm *kuberlogicv1.KuberLogicBackupRestore)
}

type InternalDetails interface {
	GetCredentialsSecret() (*v1.Secret, error)

	GetPodReplicaSelector() map[string]string
	GetPodMasterSelector() map[string]string

	GetMasterService() string
	GetReplicaService() string
	GetAccessPort() int

	GetDefaultConnectionPassword() (string, string)
	GetMainPodContainer() string
}

func GetOperator(t string) (OperatorInterface, error) {
	var operators = map[string]OperatorInterface{
		"postgresql": &postgresql.Postgres{},
		"mysql":      &mysql.Mysql{},
		//"redis":      &redis.Redis{},

	}

	value, ok := operators[t]
	if !ok {
		return nil, errors.Errorf("BaseOperator %s is not supported", t)
	}
	return value, nil
}
