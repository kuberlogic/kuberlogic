package kuberlogic

import (
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	serviceOperator "github.com/kuberlogic/kuberlogic/modules/operator/service-operator"
	"github.com/kuberlogic/kuberlogic/modules/operator/service-operator/interfaces"
	"k8s.io/client-go/kubernetes"
)

func GetCluster(cm *kuberlogicv1.KuberLogicService) (op interfaces.OperatorInterface, err error) {
	op, err = serviceOperator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)
	return
}

func GetClusterPodLabels(cm *kuberlogicv1.KuberLogicService) (master map[string]string, replica map[string]string, err error) {
	op, err := GetCluster(cm)
	if err != nil {
		return
	}
	master, replica = op.GetInternalDetails().GetPodMasterSelector(), op.GetInternalDetails().GetPodReplicaSelector()

	return
}

func GetClusterServices(cm *kuberlogicv1.KuberLogicService) (master string, replica string, err error) {
	op, err := GetCluster(cm)
	if err != nil {
		return
	}
	master, replica = op.GetInternalDetails().GetMasterService(), op.GetInternalDetails().GetReplicaService()
	return
}

func GetClusterServicePort(cm *kuberlogicv1.KuberLogicService) (p int, err error) {
	op, err := GetCluster(cm)
	if err != nil {
		return
	}
	p = op.GetInternalDetails().GetAccessPort()
	return
}

func GetClusterMainContainer(cm *kuberlogicv1.KuberLogicService) (c string, err error) {
	op, err := GetCluster(cm)
	if err != nil {
		return
	}
	c = op.GetInternalDetails().GetMainPodContainer()
	return
}

func GetClusterCredentialsInfo(cm *kuberlogicv1.KuberLogicService) (username, passwordField, secretName string, err error) {
	op, err := GetCluster(cm)
	if err != nil {
		return
	}
	details := op.GetInternalDetails()
	secretName, passwordField = details.GetDefaultConnectionPassword()
	username = details.GetDefaultConnectionUser()
	return
}

func GetSession(cm *kuberlogicv1.KuberLogicService, client *kubernetes.Clientset, db string) (session interfaces.Session, err error) {
	op, err := GetCluster(cm)
	if err != nil {
		return
	}
	session, err = op.GetSession(cm, client, db)
	return
}
