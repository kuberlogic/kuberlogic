package service_operator

import (
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/service-operator/mysql"
	"github.com/kuberlogic/operator/modules/operator/service-operator/postgresql"
	"github.com/kuberlogic/operator/modules/operator/service-operator/redis"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Operator interface {
	Name(cm *kuberlogicv1.KuberLogicService) string
	Init(cm *kuberlogicv1.KuberLogicService)
	InitFrom(o runtime.Object)
	Update(cm *kuberlogicv1.KuberLogicService)
	AsRuntimeObject() runtime.Object
	AsMetaObject() metav1.Object
	IsEqual(cm *kuberlogicv1.KuberLogicService) bool
	CurrentStatus() string
	GetDefaults() kuberlogicv1.Defaults

	GetCredentialsSecret() (*v1.Secret, error)

	GetPodReplicaSelector() map[string]string
	GetPodMasterSelector() map[string]string

	GetMasterService() string
	GetReplicaService() string
	GetAccessPort() int

	GetDefaultConnectionPassword() (string, string)
	GetMainPodContainer() string
}

type Backup interface {
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

type Restore interface {
	New(backup *kuberlogicv1.KuberLogicBackupRestore) batchv1.Job
	Init(*kuberlogicv1.KuberLogicBackupRestore)
	InitFrom(*batchv1.Job)
	GetJob() *batchv1.Job
	CurrentStatus() string

	SetRestoreImage()
	SetRestoreEnv(cm *kuberlogicv1.KuberLogicBackupRestore)
}

func GetOperator(t string) (Operator, error) {
	var operators = map[string]Operator{
		"postgresql": &postgresql.Postgres{},
		"redis":      &redis.Redis{},
		"mysql":      &mysql.Mysql{},
	}

	value, ok := operators[t]
	if !ok {
		return nil, errors.Errorf("Operator %s is not supported", t)
	}
	return value, nil
}

func GetBackupOperator(op interface{}) (Backup, error) {
	switch cluster := op.(type) {
	case *mysql.Mysql:
		return &mysql.Backup{
			Cluster: *cluster,
		}, nil
	case *postgresql.Postgres:
		return &postgresql.Backup{
			Cluster: *cluster,
		}, nil
	default:
		return nil, errors.Errorf("Cluster %s is not supported (%T)",
			cluster, cluster)
	}
}

func GetRestoreOperator(op interface{}) (Restore, error) {
	switch cluster := op.(type) {
	case *mysql.Mysql:
		return &mysql.Restore{
			Cluster: *cluster,
		}, nil
	case *postgresql.Postgres:
		return &postgresql.Restore{
			Cluster: *cluster,
		}, nil
	default:
		return nil, errors.Errorf("Cluster %s is not supported (%T)",
			cluster, cluster)
	}
}
