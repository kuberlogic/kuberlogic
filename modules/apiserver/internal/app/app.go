package app

import (
	"context"
	"github.com/kuberlogic/operator/modules/apiserver/internal/logging"
	"github.com/kuberlogic/operator/modules/apiserver/internal/security"
	"github.com/kuberlogic/operator/modules/apiserver/internal/store"
	"github.com/kuberlogic/operator/modules/apiserver/util/k8s"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Service struct {
	clientset        *kubernetes.Clientset
	kuberlogicClient *rest.RESTClient
	authProvider     *security.AuthProvider
	log              logging.Logger
	serviceStore     *store.ServiceStore
}

func (srv *Service) LookupService(ns, name string) (*kuberlogicv1.KuberLogicService, bool, error) {
	item := new(kuberlogicv1.KuberLogicService)
	err := srv.kuberlogicClient.Get().
		Namespace(ns).
		Resource("kuberlogicservices").
		Name(name).
		Do(context.TODO()).
		Into(item)

	if k8s.ErrNotFound(err) {
		return nil, false, err
	} else if err != nil {
		return nil, true, err
	}
	return item, true, nil
}

func (srv *Service) GetLogger() logging.Logger {
	return srv.log
}

func (srv *Service) GetAuthProvider() *security.AuthProvider {
	return srv.authProvider
}

func New(clientset *kubernetes.Clientset, client *rest.RESTClient, authProvider *security.AuthProvider, log logging.Logger) *Service {
	return &Service{
		clientset:        clientset,
		kuberlogicClient: client,
		authProvider:     authProvider,
		log:              log,
		serviceStore:     store.NewServiceStore(clientset, client, log),
	}
}

func (srv *Service) OnShutdown() {
	defer srv.log.Sync()
}
