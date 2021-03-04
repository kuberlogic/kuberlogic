package tests

import (
	"context"
	"github.com/kuberlogic/operator/modules/apiserver/internal/config"
	"github.com/kuberlogic/operator/modules/apiserver/internal/logging"
	k8s2 "github.com/kuberlogic/operator/modules/apiserver/util/k8s"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/watcher/k8s"
	"k8s.io/client-go/kubernetes"
)

func Connect(ns, name string) (*kubernetes.Clientset, *kuberlogicv1.KuberLogicService, error) {
	internalCfg, err := config.InitConfig("kuberlogic", logging.WithComponentLogger("config"))
	if err != nil {
		return nil, nil, err
	}

	cfg, err := k8s2.GetConfig(internalCfg)
	if err != nil {
		return nil, nil, err
	}

	client, err := k8s.GetBaseClient(cfg)
	if err != nil {
		return nil, nil, err
	}

	crdClient, err := k8s.GetKuberLogicClient(cfg)
	if err != nil {
		return nil, nil, err
	}
	cluster := &kuberlogicv1.KuberLogicService{}
	err = crdClient.
		Get().
		Resource("kuberlogicservices").
		Namespace(ns).
		Name(name).
		Do(context.TODO()).
		Into(cluster)
	if err != nil {
		return nil, nil, err
	}
	return client, cluster, err
}
