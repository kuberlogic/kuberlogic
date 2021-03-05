package mysql

import (
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/service-operator/base"
	"github.com/kuberlogic/operator/modules/operator/service-operator/interfaces"
	"github.com/kuberlogic/operator/modules/operator/util"
	mysqlv1 "github.com/presslabs/mysql-operator/pkg/apis/mysql/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	image   = "mysql"
	version = "5.7.31"

	mysqlRoleKey     = "role"
	mysqlRoleReplica = "replica"
	mysqlRoleMaster  = "master"

	mysqlPodLabelKey   = "mysql.presslabs.org/cluster"
	mysqlMainContainer = "mysql"

	mysqlPort = 3306
)

type Mysql struct {
	base.BaseOperator
	Operator *mysqlv1.MysqlCluster
}

func (p *Mysql) GetBackupSchedule() interfaces.BackupSchedule {
	return &Backup{
		Cluster: p,
	}
}

func (p *Mysql) GetBackupRestore() interfaces.BackupRestore {
	return &Restore{
		Cluster: p,
	}
}

func (p *Mysql) GetInternalDetails() interfaces.InternalDetails {
	return &InternalDetails{
		Cluster: p,
	}
}

func (p *Mysql) Init(cm *kuberlogicv1.KuberLogicService) {
	p.Operator = &mysqlv1.MysqlCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Name(cm),
			Namespace: cm.Namespace,
		},
		Spec: mysqlv1.MysqlClusterSpec{
			SecretName: genCredentialsSecretName(cm.Name),
			PodSpec: mysqlv1.PodSpec{
				Annotations: map[string]string{
					"monitoring.cloudlinux.com/scrape": "true",
					"monitoring.cloudlinux.com/port":   "9999",
				},
				Containers: []corev1.Container{
					{
						Name:  "kuberlogic-exporter",
						Image: "quay.io/kuberlogic/mysql-exporter-deprecated:v2",
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
	mysqlv1.SetDefaults_MysqlCluster(p.Operator)
}

func (p *Mysql) InitFrom(o runtime.Object) {
	p.Operator = o.(*mysqlv1.MysqlCluster)
}

func (p *Mysql) GetDefaults() kuberlogicv1.Defaults {
	return kuberlogicv1.Defaults{
		VolumeSize: kuberlogicv1.DefaultVolumeSize,
		Resources:  kuberlogicv1.DefaultResources,
		Version:    version,
	}
}

func (p *Mysql) Update(cm *kuberlogicv1.KuberLogicService) {
	p.setReplica(cm)
	p.setResources(cm)
	p.setVolumeSize(cm)
	p.setImage(cm)
	p.setAdvancedConf(cm)
}

func (p *Mysql) setReplica(cm *kuberlogicv1.KuberLogicService) {
	p.Operator.Spec.Replicas = &cm.Spec.Replicas
}

func (p *Mysql) setResources(cm *kuberlogicv1.KuberLogicService) {
	p.Operator.Spec.PodSpec.Resources = cm.Spec.Resources
}

func (p *Mysql) setVolumeSize(cm *kuberlogicv1.KuberLogicService) {
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

func (p *Mysql) setImage(cm *kuberlogicv1.KuberLogicService) {
	p.Operator.Spec.Image = util.GetImage(image, cm.Spec.Version)

	secrets := []corev1.LocalObjectReference{
		{Name: util.GetImagePullSecret()},
	}
	p.Operator.Spec.PodSpec.ImagePullSecrets = secrets
}

func (p *Mysql) setAdvancedConf(cm *kuberlogicv1.KuberLogicService) {
	desiredMysqlConf := util.StrToIntOrStr(cm.Spec.AdvancedConf)

	if p.Operator.Spec.MysqlConf == nil {
		p.Operator.Spec.MysqlConf = make(map[string]intstr.IntOrString)
	}

	for k, v := range desiredMysqlConf {
		p.Operator.Spec.MysqlConf[k] = v
	}
}

func (p *Mysql) IsEqual(cm *kuberlogicv1.KuberLogicService) bool {
	return p.isEqualReplica(cm) &&
		p.isEqualResources(cm) &&
		p.isEqualVolumeSize(cm) &&
		p.isEqualImage(cm) &&
		p.isEqualAdvancedConf(cm)
}

func (p *Mysql) isEqualReplica(cm *kuberlogicv1.KuberLogicService) bool {
	return *p.Operator.Spec.Replicas == cm.Spec.Replicas
}

func (p *Mysql) isEqualResources(cm *kuberlogicv1.KuberLogicService) bool {
	op := p.Operator.Spec.PodSpec.Resources
	cmr := cm.Spec.Resources
	return op.Limits.Cpu().Cmp(*cmr.Limits.Cpu()) == 0 &&
		op.Limits.Memory().Cmp(*cmr.Limits.Memory()) == 0 &&
		op.Requests.Cpu().Cmp(*cmr.Requests.Cpu()) == 0 &&
		op.Requests.Memory().Cmp(*cmr.Requests.Memory()) == 0
}

func (p *Mysql) isEqualVolumeSize(cm *kuberlogicv1.KuberLogicService) bool {
	if &p.Operator.Spec.VolumeSpec.PersistentVolumeClaim.Resources == nil {
		return false
	}
	return p.Operator.Spec.VolumeSpec.PersistentVolumeClaim.Resources.Requests.Storage().Cmp(
		resource.MustParse(cm.Spec.VolumeSize),
	) == 0
}

func (p *Mysql) isEqualImage(cm *kuberlogicv1.KuberLogicService) bool {
	return p.Operator.Spec.Image == util.GetImage(image, cm.Spec.Version)
}

func (p *Mysql) isEqualAdvancedConf(cm *kuberlogicv1.KuberLogicService) bool {
	desiredMysqlConf := util.StrToIntOrStr(cm.Spec.AdvancedConf)
	for k, v := range desiredMysqlConf {
		if val, ok := p.Operator.Spec.MysqlConf[k]; !ok {
			return false
		} else if val != v {
			return false
		}
	}
	return true
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
		return kuberlogicv1.ClusterNotReadyStatus
	case "True":
		return kuberlogicv1.ClusterOkStatus
	default:
		return kuberlogicv1.ClusterUnknownStatus
	}
}

func genCredentialsSecretName(cluster string) string {
	return cluster + "-cred"
}
