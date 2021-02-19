package redis

import (
	redisv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	kuberlogicv1 "gitlab.com/cloudmanaged/operator/api/v1"
	"gitlab.com/cloudmanaged/operator/api/v1/operator/util"
	"k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const baseRedisImage = "redis"
const latestRedisVersion = "6.0.8"

const (
	redisRoleKey     = ""
	redisRoleReplica = ""
	redisRoleMaster  = ""
)

type Redis struct {
	Operator redisv1.RedisFailover
}

func (p *Redis) Name(cm *kuberlogicv1.KuberLogicService) string {
	return cm.Name
}

func (p *Redis) AsRuntimeObject() runtime.Object {
	return &p.Operator
}

func (p *Redis) AsMetaObject() metav1.Object {
	return &p.Operator
}

func (p *Redis) Init(cm *kuberlogicv1.KuberLogicService) {
	p.Operator.ObjectMeta = metav1.ObjectMeta{
		Name:      cm.Name,
		Namespace: cm.Namespace,
	}
}

func (p *Redis) InitFrom(o runtime.Object) {
	p.Operator = *o.(*redisv1.RedisFailover)
}

func (p *Redis) GetDefaults() kuberlogicv1.Defaults {
	return kuberlogicv1.Defaults{
		VolumeSize: kuberlogicv1.DefaultVolumeSize,
		Resources:  kuberlogicv1.DefaultResources,
		Version:    latestRedisVersion,
	}
}

func (p *Redis) Update(cm *kuberlogicv1.KuberLogicService) {
	p.setReplica(cm)
	p.setResources(cm)
	p.setVolumeSize(cm)
	p.setImage(cm)
}

func (p *Redis) setReplica(cm *kuberlogicv1.KuberLogicService) {
	p.Operator.Spec.Redis.Replicas = cm.Spec.Replicas
}

func (p *Redis) setResources(cm *kuberlogicv1.KuberLogicService) {
	if &cm.Spec.Resources != nil {
		p.Operator.Spec.Redis.Resources = cm.Spec.Resources
		p.Operator.Spec.Sentinel.Resources = cm.Spec.Resources
	}
}

func (p *Redis) setVolumeSize(cm *kuberlogicv1.KuberLogicService) {
	p.Operator.Spec.Redis.Storage.KeepAfterDeletion = true

	resources := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceStorage: resource.MustParse(cm.Spec.VolumeSize),
		},
	}
	if p.Operator.Spec.Redis.Storage.PersistentVolumeClaim == nil {
		p.Operator.Spec.Redis.Storage.PersistentVolumeClaim = &v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: cm.Name + "-data",
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: []v1.PersistentVolumeAccessMode{
					v1.ReadWriteOnce,
				},
				Resources: resources,
			},
		}
	} else {
		p.Operator.Spec.Redis.Storage.PersistentVolumeClaim.Spec.Resources = resources
	}
}

func (p *Redis) setImage(cm *kuberlogicv1.KuberLogicService) {
	p.Operator.Spec.Redis.Image = util.GetImage(baseRedisImage, cm.Spec.Version)
	p.Operator.Spec.Sentinel.Image = util.GetImage(baseRedisImage, cm.Spec.Version)

	secrets := []v1.LocalObjectReference{
		{Name: util.GetImagePullSecret()},
	}
	p.Operator.Spec.Redis.ImagePullSecrets = secrets
	p.Operator.Spec.Sentinel.ImagePullSecrets = secrets
}

func (p *Redis) IsEqual(cm *kuberlogicv1.KuberLogicService) bool {
	return p.isEqualReplica(cm) &&
		p.isEqualResources(cm) &&
		p.isEqualVolumeSize(cm) &&
		p.isEqualImage(cm)
}

func (p *Redis) isEqualReplica(cm *kuberlogicv1.KuberLogicService) bool {
	return p.Operator.Spec.Redis.Replicas == cm.Spec.Replicas
}

func (p *Redis) isEqualResources(cm *kuberlogicv1.KuberLogicService) bool {
	op := p.Operator.Spec.Redis.Resources
	cmr := cm.Spec.Resources
	return op.Limits.Cpu().Cmp(*cmr.Limits.Cpu()) == 0 &&
		op.Limits.Memory().Cmp(*cmr.Limits.Memory()) == 0 &&
		op.Requests.Cpu().Cmp(*cmr.Requests.Cpu()) == 0 &&
		op.Requests.Memory().Cmp(*cmr.Requests.Memory()) == 0
}

func (p *Redis) isEqualVolumeSize(cm *kuberlogicv1.KuberLogicService) bool {
	if &p.Operator.Spec.Redis.Storage.PersistentVolumeClaim == nil {
		return false
	}
	return p.Operator.Spec.Redis.Storage.PersistentVolumeClaim.Spec.Resources.Requests.Storage().Cmp(
		resource.MustParse(cm.Spec.VolumeSize),
	) == 0
}

func (p *Redis) isEqualImage(cm *kuberlogicv1.KuberLogicService) bool {
	image := util.GetImage(baseRedisImage, cm.Spec.Version)
	return p.Operator.Spec.Redis.Image == image && p.Operator.Spec.Sentinel.Image == image
}

func (p *Redis) CurrentStatus() string {
	// TODO: task for implementation https://gitlab.corp.cloudlinux.com/cloudmanaged/cloudmanaged/-/issues/17
	return ""
}

func (p *Redis) GenerateJob(backup *kuberlogicv1.KuberLogicBackupSchedule) v1beta1.CronJob {
	return v1beta1.CronJob{}
}

func (p *Redis) GetPodReplicaSelector() map[string]string {
	return map[string]string{redisRoleKey: redisRoleReplica}
}

func (p *Redis) GetPodMasterSelector() map[string]string {
	return map[string]string{redisRoleKey: redisRoleMaster}
}

func (p *Redis) GetMasterService() string {
	return ""
}

func (p *Redis) GetReplicaService() string {
	return ""
}

func (p *Redis) GetAccessPort() int {
	return 0
}

func (p *Redis) GetMainPodContainer() string {
	return ""
}

func (p *Redis) GetDefaultConnectionPassword() (string, string) {
	return "", ""
}

func (p *Redis) GetCredentialsSecret() (*v1.Secret, error) {
	return nil, nil
}
