package cloudmanaged

import (
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	"gitlab.com/cloudmanaged/operator/api/v1/operator"
)

func GetClusterPodLabels(cm *cloudlinuxv1.CloudManaged) (master map[string]string, replica map[string]string, err error) {
	op, err := operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}
	master, replica = op.GetPodMasterSelector(cm.Name), op.GetPodReplicaSelector(cm.Name)

	return
}

func GetClusterServices(cm *cloudlinuxv1.CloudManaged) (master string, replica string, err error) {
	op, err := operator.GetOperator(cm.Spec.Type)
	if err != nil {
		return
	}

	master, replica = op.GetMasterService(cm.Name, cm.Namespace), op.GetReplicaService(cm.Name, cm.Namespace)
	return
}

func GetClusterServicePort(cm *cloudlinuxv1.CloudManaged) (p int, e error) {
	op, e := operator.GetOperator(cm.Spec.Type)
	if e != nil {
		return
	}

	p = op.GetAccessPort()
	return
}
