package base

import (
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Op interface {
	metav1.Object
	runtime.Object
}

type BaseOperator struct {
	Operator Op

	backup  BaseBackup
	restore BaseRestore
}

func (p *BaseOperator) Name(cm *kuberlogicv1.KuberLogicService) string {
	return cm.Name
}

func (p *BaseOperator) AsRuntimeObject() runtime.Object {
	return p.Operator
}

func (p *BaseOperator) AsMetaObject() metav1.Object {
	return p.Operator
}
