package kuberlogic

import (
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/service-operator"
)

func GetClusterName(cm *kuberlogicv1.KuberLogicService) (name string, err error) {
	op, err := service_operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)

	return op.Name(cm), nil
}

func GetClusterPodLabels(cm *kuberlogicv1.KuberLogicService) (master map[string]string, replica map[string]string, err error) {
	op, err := service_operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)
	master, replica = op.GetInternalDetails().GetPodMasterSelector(), op.GetInternalDetails().GetPodReplicaSelector()

	return
}

func GetClusterServices(cm *kuberlogicv1.KuberLogicService) (master string, replica string, err error) {
	op, err := service_operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)

	master, replica = op.GetInternalDetails().GetMasterService(), op.GetInternalDetails().GetReplicaService()
	return
}

func GetClusterServicePort(cm *kuberlogicv1.KuberLogicService) (p int, err error) {
	op, err := service_operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)

	p = op.GetInternalDetails().GetAccessPort()
	return
}

func GetClusterMainContainer(cm *kuberlogicv1.KuberLogicService) (c string, err error) {
	op, err := service_operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)

	c = op.GetInternalDetails().GetMainPodContainer()
	return
}

func GetClusterCredentialsInfo(cm *kuberlogicv1.KuberLogicService) (username, passwordField, secretName string, err error) {
	op, err := service_operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)

	secretName, passwordField = op.GetInternalDetails().GetDefaultConnectionPassword()
	username = kuberlogicv1.DefaultUser
	return
}
