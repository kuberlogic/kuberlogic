package api

import (
	"github.com/pkg/errors"
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	"gitlab.com/cloudmanaged/operator/watcher/api/common"
	"gitlab.com/cloudmanaged/operator/watcher/api/mysql"
	"gitlab.com/cloudmanaged/operator/watcher/api/postgres"
	"k8s.io/client-go/kubernetes"
)

func GetSession(cm *cloudlinuxv1.CloudManaged, client *kubernetes.Clientset, db, table string) (common.Session, error) {
	switch cm.Spec.Type {
	case "mysql":
		return mysql.New(cm, client, db, table)
	case "postgresql":
		return postgres.New(cm, client, db, table)
	default:
		return nil, errors.Errorf("Cluster %s is not supported", cm.Spec.Type)
	}

}
