package app

import (
	"github.com/kuberlogic/operator/modules/apiserver/internal/logging"
	"github.com/kuberlogic/operator/modules/apiserver/internal/security"
	"github.com/kuberlogic/operator/modules/apiserver/internal/store"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Service struct {
	clientset    *kubernetes.Clientset
	cmClient     *rest.RESTClient
	authProvider security.AuthProvider
	log          logging.Logger

	serviceStore *store.ServiceStore
}

func New(clientset *kubernetes.Clientset, client *rest.RESTClient, authProvider security.AuthProvider, log logging.Logger) *Service {
	return &Service{
		clientset:    clientset,
		cmClient:     client,
		authProvider: authProvider,
		log:          log,
		serviceStore: store.NewServiceStore(clientset, client, log),
	}
}

func (srv *Service) OnShutdown() {
	defer srv.log.Sync()
}

// TODO:
// + in-cluster setup (add into operators pod)
// +/- add resources/limits, versions, volumeSize for the cloudmanged CR
// + add statuses of CR for the list commands
// + figure out how to generate generic endpoint for the error messages
// - add ability to edit credentials for backups/restores
// - check crud + (backup/restore) for mysql (need to create additional secret)
// - set owner references for the related secrets/resources
// - integration tests
// - rename backup -> scheduledBackup
