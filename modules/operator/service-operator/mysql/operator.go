/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mysql

import (
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	"github.com/kuberlogic/kuberlogic/modules/operator/service-operator/interfaces"
	"github.com/kuberlogic/kuberlogic/modules/operator/util"
	mysqlv1 "github.com/presslabs/mysql-operator/pkg/apis/mysql/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	image = "mysql"
)

type Mysql struct {
	Operator mysqlv1.MysqlCluster
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

func (p *Mysql) GetSession(kls *kuberlogicv1.KuberLogicService, client kubernetes.Interface, db string) (interfaces.Session, error) {
	return NewSession(p, kls, client, db)
}

func (p *Mysql) Name(kls *kuberlogicv1.KuberLogicService) string {
	return kls.Name
}

func (p *Mysql) AsRuntimeObject() runtime.Object {
	return &p.Operator
}

func (p *Mysql) AsMetaObject() metav1.Object {
	return &p.Operator
}

func (p *Mysql) AsClientObject() client.Object {
	return &p.Operator
}

func (p *Mysql) Init(kls *kuberlogicv1.KuberLogicService) {
	p.Operator = mysqlv1.MysqlCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Name(kls),
			Namespace: kls.Namespace,
		},
		Spec: mysqlv1.MysqlClusterSpec{
			SecretName:   genCredentialsSecretName(kls.Name),
			MysqlVersion: kls.Spec.Version,
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
						Image: util.GetKuberlogicImage(image, kls.Spec.Version),
						Command: []string{
							"/bin/sh",
							"-c",
							`
MYSQL_DIR=/var/lib/mysql/mysql
if [ -d $MYSQL_DIR ] 
then
	for f in $(ls $MYSQL_DIR/*MYI); do 
		myisamchk -r --update-state $(echo $f | tr -d .MYI); 
	done
else
	echo "Directory $MYSQL_DIR does not exists"
fi
`,
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

func (p *Mysql) Update(cm *kuberlogicv1.KuberLogicService) {
	p.setReplica(cm)
	p.setResources(cm)
	p.setVolumeSize(cm)
	p.setImage(cm)
	p.setAdvancedConf(cm)
}

func (p *Mysql) setReplica(kls *kuberlogicv1.KuberLogicService) {
	p.Operator.Spec.Replicas = &kls.Spec.Replicas
}

func (p *Mysql) setResources(kls *kuberlogicv1.KuberLogicService) {
	p.Operator.Spec.PodSpec.Resources = kls.Spec.Resources
}

func (p *Mysql) setVolumeSize(kls *kuberlogicv1.KuberLogicService) {
	resources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceStorage: resource.MustParse(kls.Spec.VolumeSize),
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

func (p *Mysql) setImage(kls *kuberlogicv1.KuberLogicService) {
	p.Operator.Spec.Image = util.GetKuberlogicImage(image, kls.Spec.Version)

	if util.GetKuberlogicRepoPullSecret() != "" {
		secrets := []corev1.LocalObjectReference{
			{Name: util.GetKuberlogicRepoPullSecret()},
		}
		p.Operator.Spec.PodSpec.ImagePullSecrets = secrets
	}
}

func (p *Mysql) setAdvancedConf(kls *kuberlogicv1.KuberLogicService) {
	desiredMysqlConf := util.StrToIntOrStr(kls.Spec.AdvancedConf)

	if p.Operator.Spec.MysqlConf == nil {
		p.Operator.Spec.MysqlConf = make(map[string]intstr.IntOrString)
	}

	for k, v := range desiredMysqlConf {
		p.Operator.Spec.MysqlConf[k] = v
	}
}

func (p *Mysql) IsEqual(kls *kuberlogicv1.KuberLogicService) bool {
	return p.isEqualReplica(kls) &&
		p.isEqualResources(kls) &&
		p.isEqualVolumeSize(kls) &&
		p.isEqualImage(kls) &&
		p.isEqualAdvancedConf(kls)
}

func (p *Mysql) isEqualReplica(kls *kuberlogicv1.KuberLogicService) bool {
	return *p.Operator.Spec.Replicas == kls.Spec.Replicas
}

func (p *Mysql) isEqualResources(kls *kuberlogicv1.KuberLogicService) bool {
	op := p.Operator.Spec.PodSpec.Resources
	cmr := kls.Spec.Resources
	return op.Limits.Cpu().Cmp(*cmr.Limits.Cpu()) == 0 &&
		op.Limits.Memory().Cmp(*cmr.Limits.Memory()) == 0 &&
		op.Requests.Cpu().Cmp(*cmr.Requests.Cpu()) == 0 &&
		op.Requests.Memory().Cmp(*cmr.Requests.Memory()) == 0
}

func (p *Mysql) isEqualVolumeSize(kls *kuberlogicv1.KuberLogicService) bool {
	if &p.Operator.Spec.VolumeSpec.PersistentVolumeClaim.Resources == nil {
		return false
	}
	return p.Operator.Spec.VolumeSpec.PersistentVolumeClaim.Resources.Requests.Storage().Cmp(
		resource.MustParse(kls.Spec.VolumeSize),
	) == 0
}

func (p *Mysql) isEqualImage(kls *kuberlogicv1.KuberLogicService) bool {
	return p.Operator.Spec.Image == util.GetKuberlogicImage(image, kls.Spec.Version)
}

func (p *Mysql) isEqualAdvancedConf(kls *kuberlogicv1.KuberLogicService) bool {
	desiredMysqlConf := util.StrToIntOrStr(kls.Spec.AdvancedConf)
	for k, v := range desiredMysqlConf {
		if val, ok := p.Operator.Spec.MysqlConf[k]; !ok {
			return false
		} else if val != v {
			return false
		}
	}
	return true
}

func (p *Mysql) IsReady() (bool, string) {
	status := ""
	for _, v := range p.Operator.Status.Conditions {
		if v.Type == "Ready" {
			status = string(v.Status)
		}
	}

	switch status {
	case "False":
		return false, kuberlogicv1.ClusterNotReadyStatus
	case "True":
		return true, kuberlogicv1.ClusterOkStatus
	default:
		return false, kuberlogicv1.ClusterUnknownStatus
	}
}

func genCredentialsSecretName(cluster string) string {
	return cluster + "-cred"
}
