package kuberlogic

import (
	kuberlogicv1 "gitlab.com/cloudmanaged/operator/api/v1"
	"gitlab.com/cloudmanaged/operator/api/v1/operator"
)

func GetClusterName(cm *kuberlogicv1.KuberLogicService) (name string, err error) {
	op, err := operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)

	return op.Name(cm), nil
}

func GetClusterPodLabels(cm *kuberlogicv1.KuberLogicService) (master map[string]string, replica map[string]string, err error) {
	op, err := operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)
	master, replica = op.GetPodMasterSelector(), op.GetPodReplicaSelector()

	return
}

func GetClusterServices(cm *kuberlogicv1.KuberLogicService) (master string, replica string, err error) {
	op, err := operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)

	master, replica = op.GetMasterService(), op.GetReplicaService()
	return
}

func GetClusterServicePort(cm *kuberlogicv1.KuberLogicService) (p int, err error) {
	op, err := operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)

	p = op.GetAccessPort()
	return
}

func GetClusterMainContainer(cm *kuberlogicv1.KuberLogicService) (c string, err error) {
	op, err := operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)

	c = op.GetMainPodContainer()
	return
}

func GetClusterCredentialsInfo(cm *kuberlogicv1.KuberLogicService) (username, passwordField, secretName string, err error) {
	op, err := operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)

	secretName, passwordField = op.GetDefaultConnectionPassword()
	username = kuberlogicv1.DefaultUser
	return
}
