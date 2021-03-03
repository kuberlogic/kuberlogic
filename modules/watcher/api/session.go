package api

import (
	kuberlogicv1 "github.com/kuberlogic/operator/pkg/operator/api/v1"
	"github.com/kuberlogic/operator/pkg/watcher/api/common"
	"github.com/kuberlogic/operator/pkg/watcher/api/mysql"
	"github.com/kuberlogic/operator/pkg/watcher/api/postgres"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
)

func GetSession(cm *kuberlogicv1.KuberLogicService, client *kubernetes.Clientset, db, table string) (common.Session, error) {
	switch cm.Spec.Type {
	case "mysql":
		return mysql.New(cm, client, db, table)
	case "postgresql":
		return postgres.New(cm, client, db, table)
	default:
		return nil, errors.Errorf("Cluster %s is not supported", cm.Spec.Type)
	}

}
