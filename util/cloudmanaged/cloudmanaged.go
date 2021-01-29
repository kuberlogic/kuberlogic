package cloudmanaged

import (
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	"gitlab.com/cloudmanaged/operator/api/v1/operator"
)

func GetClusterName(cm *cloudlinuxv1.CloudManaged) (name string, err error) {
	op, err := operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)

	return op.Name(cm), nil
}

func GetClusterPodLabels(cm *cloudlinuxv1.CloudManaged) (master map[string]string, replica map[string]string, err error) {
	op, err := operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)
	master, replica = op.GetPodMasterSelector(), op.GetPodReplicaSelector()

	return
}

func GetClusterServices(cm *cloudlinuxv1.CloudManaged) (master string, replica string, err error) {
	op, err := operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)

	master, replica = op.GetMasterService(), op.GetReplicaService()
	return
}

func GetClusterServicePort(cm *cloudlinuxv1.CloudManaged) (p int, err error) {
	op, err := operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)

	p = op.GetAccessPort()
	return
}

func GetClusterMainContainer(cm *cloudlinuxv1.CloudManaged) (c string, err error) {
	op, err := operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)

	c = op.GetMainPodContainer()
	return
}

func GetClusterCredentialsInfo(cm *cloudlinuxv1.CloudManaged) (username, passwordField, secretName string, err error) {
	op, err := operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	op.Init(cm)

	secretName, passwordField = op.GetDefaultConnectionPassword()
	username = cm.Spec.DefaultUser
	return
}
