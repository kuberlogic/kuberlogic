package store

import (
	"github.com/kuberlogic/operator/modules/apiserver/internal/logging"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ServiceStore struct {
	restClient *rest.RESTClient
	clientset  *kubernetes.Clientset
	log        logging.Logger
}
