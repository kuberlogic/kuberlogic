package operator

import (
	mysqlv1 "github.com/presslabs/mysql-operator/pkg/apis/mysql/v1alpha1"
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const baseMysqlImage = "percona"
const latestMysqlVersion = "5.7.31"

type Mysql struct {
	Operator mysqlv1.MysqlCluster
}

func (p *Mysql) AsRuntimeObject() runtime.Object {
	return &p.Operator
}

func (p *Mysql) AsMetaObject() metav1.Object {
	return &p.Operator
}

func (p *Mysql) Init(cm *cloudlinuxv1.CloudManaged) {
	p.Operator = mysqlv1.MysqlCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cm.Name,
			Namespace: cm.Namespace,
		},
	}
	mysqlv1.SetDefaults_MysqlCluster(&p.Operator)
}

func (p *Mysql) InitFrom(o runtime.Object) {
	p.Operator = *o.(*mysqlv1.MysqlCluster)
}

func (p *Mysql) GetDefaults() cloudlinuxv1.Defaults {
	return cloudlinuxv1.Defaults{
		VolumeSize: cloudlinuxv1.DefaultVolumeSize,
		Resources:  cloudlinuxv1.DefaultResources,
		Version:    latestMysqlVersion,
	}
}

func (p *Mysql) Update(cm *cloudlinuxv1.CloudManaged) {
	p.setReplica(cm)
	p.setSecret(cm)
	p.setResources(cm)
	p.setVolumeSize(cm)
	p.setImage(cm)
}

func (p *Mysql) setReplica(cm *cloudlinuxv1.CloudManaged) {
	p.Operator.Spec.Replicas = &cm.Spec.Replicas
}

func (p *Mysql) setSecret(cm *cloudlinuxv1.CloudManaged) {
	p.Operator.Spec.SecretName = cm.Spec.SecretName
}

func (p *Mysql) setResources(cm *cloudlinuxv1.CloudManaged) {
	p.Operator.Spec.PodSpec.Resources = cm.Spec.Resources
}

func (p *Mysql) setVolumeSize(cm *cloudlinuxv1.CloudManaged) {
	resources := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceStorage: resource.MustParse(cm.Spec.VolumeSize),
		},
	}
	if p.Operator.Spec.VolumeSpec.PersistentVolumeClaim == nil {
		p.Operator.Spec.VolumeSpec.PersistentVolumeClaim = &v1.PersistentVolumeClaimSpec{
			Resources: resources,
		}
	} else {
		p.Operator.Spec.VolumeSpec.PersistentVolumeClaim.Resources = resources
	}
}

func (p *Mysql) setImage(cm *cloudlinuxv1.CloudManaged) {
	p.Operator.Spec.Image = getImage(baseMysqlImage, cm.Spec.Version)

	secrets := []v1.LocalObjectReference{
		{Name: getImagePullSecret()},
	}
	p.Operator.Spec.PodSpec.ImagePullSecrets = secrets
}

func (p *Mysql) IsEqual(cm *cloudlinuxv1.CloudManaged) bool {
	return p.isEqualReplica(cm) &&
		p.isEqualResources(cm) &&
		p.isEqualVolumeSize(cm) &&
		p.isEqualImage(cm)
}

func (p *Mysql) isEqualReplica(cm *cloudlinuxv1.CloudManaged) bool {
	return *p.Operator.Spec.Replicas == cm.Spec.Replicas
}

func (p *Mysql) isEqualResources(cm *cloudlinuxv1.CloudManaged) bool {
	op := p.Operator.Spec.PodSpec.Resources
	cmr := cm.Spec.Resources
	return op.Limits.Cpu().Cmp(*cmr.Limits.Cpu()) == 0 &&
		op.Limits.Memory().Cmp(*cmr.Limits.Memory()) == 0 &&
		op.Requests.Cpu().Cmp(*cmr.Requests.Cpu()) == 0 &&
		op.Requests.Memory().Cmp(*cmr.Requests.Memory()) == 0
}

func (p *Mysql) isEqualVolumeSize(cm *cloudlinuxv1.CloudManaged) bool {
	if &p.Operator.Spec.VolumeSpec.PersistentVolumeClaim.Resources == nil {
		return false
	}
	return p.Operator.Spec.VolumeSpec.PersistentVolumeClaim.Resources.Requests.Storage().Cmp(
		resource.MustParse(cm.Spec.VolumeSize),
	) == 0
}

func (p *Mysql) isEqualImage(cm *cloudlinuxv1.CloudManaged) bool {
	return p.Operator.Spec.Image == getImage(baseMysqlImage, cm.Spec.Version)
}

func (p *Mysql) CurrentStatus() string {
	if int32(p.Operator.Status.ReadyNodes) == *p.Operator.Spec.Replicas {
		return string(v1.ConditionTrue)
	} else {
		return string(v1.ConditionFalse)
	}
}
