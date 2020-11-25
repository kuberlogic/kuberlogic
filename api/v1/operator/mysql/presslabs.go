package mysql

import (
	mysqlv1 "github.com/presslabs/mysql-operator/pkg/apis/mysql/v1alpha1"
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	"gitlab.com/cloudmanaged/operator/api/v1/operator/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	image   = "mysql"
	version = "5.7.31"
)

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
		Spec: mysqlv1.MysqlClusterSpec{
			PodSpec: mysqlv1.PodSpec{
				Annotations: map[string]string{
					"monitoring.cloudlinux.com/scrape": "true",
					"monitoring.cloudlinux.com/port":   "9999",
				},
				Containers: []corev1.Container{
					{
						Name:  "cloudmanaged-exporter",
						Image: "gitlab.corp.cloudlinux.com:5001/cloudmanaged/cloudmanaged/exporter:v1",
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "data",
								MountPath: "/var/lib/mysql",
							},
						},
						Ports: []corev1.ContainerPort{
							{
								Name:          "metrics",
								ContainerPort: 9999,
								Protocol:      corev1.ProtocolTCP,
							},
						},
					},
				},
				InitContainers: []corev1.Container{
					{
						Name:  "myisam-repair",
						Image: util.GetImage(image, cm.Spec.Version),
						Command: []string{
							"/bin/sh",
							"-c",
							"for f in $(ls /var/lib/mysql/mysql/*MYI); do myisamchk -r --update-state $(echo $f | tr -d .MYI); done",
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "data",
								MountPath: "/var/lib/mysql",
							},
						},
					},
				},
			},
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
		Version:    version,
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
	resources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceStorage: resource.MustParse(cm.Spec.VolumeSize),
		},
	}
	if p.Operator.Spec.VolumeSpec.PersistentVolumeClaim == nil {
		p.Operator.Spec.VolumeSpec.PersistentVolumeClaim = &corev1.PersistentVolumeClaimSpec{
			Resources: resources,
		}
	} else {
		p.Operator.Spec.VolumeSpec.PersistentVolumeClaim.Resources = resources
	}
}

func (p *Mysql) setImage(cm *cloudlinuxv1.CloudManaged) {
	p.Operator.Spec.Image = util.GetImage(image, cm.Spec.Version)

	secrets := []corev1.LocalObjectReference{
		{Name: util.GetImagePullSecret()},
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
	return p.Operator.Spec.Image == util.GetImage(image, cm.Spec.Version)
}

func (p *Mysql) CurrentStatus() string {
	status := ""
	for _, v := range p.Operator.Status.Conditions {
		if v.Type == "Ready" {
			status = string(v.Status)
		}
	}

	switch status {
	case "False":
		return cloudlinuxv1.ClusterNotReadyStatus
	case "True":
		return cloudlinuxv1.ClusterOkStatus
	default:
		return cloudlinuxv1.ClusterUnknownStatus
	}
}
