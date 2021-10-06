package tests

import (
	"context"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/config"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/logging"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	"github.com/kuberlogic/kuberlogic/modules/operator/util"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getConfig() (*rest.Config, error) {
	// check in-cluster usage
	if cfg, err := rest.InClusterConfig(); err == nil {
		return cfg, nil
	}

	internalCfg, err := config.InitConfig("kuberlogic", logging.WithComponentLogger("config"))
	if err != nil {
		return nil, err
	}

	// use the current context in kubeconfig
	conf, err := clientcmd.BuildConfigFromFlags("", internalCfg.KubeconfigPath)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func Connect(ns, name string) (*kubernetes.Clientset, *kuberlogicv1.KuberLogicService, error) {
	cfg, err := getConfig()
	if err != nil {
		return nil, nil, err
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	crdClient, err := util.GetKuberLogicClient(cfg)
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
