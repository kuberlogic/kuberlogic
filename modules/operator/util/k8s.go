package util

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/go-errors/errors"
	"github.com/go-logr/logr"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	k8scheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"time"
)

func GetKuberLogicClient(config *rest.Config) (*rest.RESTClient, error) {
	crdConfig := *config
	crdConfig.ContentConfig.GroupVersion = &kuberlogicv1.GroupVersion
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(k8scheme.Scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	return rest.UnversionedRESTClientFor(&crdConfig)
}

func HandlePanic(log logr.Logger) {
	if err := recover(); err != nil {
		log.Error(errors.New("handle panic"), fmt.Sprintf("%v", err))
		result := sentry.Flush(5 * time.Second)
		if !result {
			time.Sleep(5 * time.Second)
		}
		panic(err)
	}
}
