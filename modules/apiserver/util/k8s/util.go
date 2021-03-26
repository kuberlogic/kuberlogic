package k8s

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kuberlogic/operator/modules/apiserver/internal/config"
	"github.com/kuberlogic/operator/modules/apiserver/internal/logging"
	"io"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
)

func GetConfig(cfg *config.Config) (*rest.Config, error) {
	// check in-cluster usage
	if cfg, err := rest.InClusterConfig(); err == nil {
		return cfg, nil
	}

	// use the current context in kubeconfig
	conf, err := clientcmd.BuildConfigFromFlags("",
		cfg.KubeconfigPath)
	if err != nil {
		return nil, err
	}
	return conf, err
}

func ErrNotFound(err error) (notFoundErr bool) {
	if statusError, isStatus := err.(*errors.StatusError); isStatus && statusError.Status().Reason == metav1.StatusReasonNotFound {
		notFoundErr = true
	}
	return
}

func MapToStrSelector(labels map[string]string) string {
	b := new(bytes.Buffer)

	for k, v := range labels {
		fmt.Fprintf(b, "%s=%s,", k, v)
	}

	return strings.TrimSuffix(b.String(), ",")
}

func GetPodLogs(c *kubernetes.Clientset, log logging.Logger, name, container, ns string, lines int64) (logs string, err error) {
	podLogOptions := v1.PodLogOptions{
		Follow:    false,
		TailLines: &lines,
		Container: container,
	}
	r := c.CoreV1().Pods(ns).GetLogs(name, &podLogOptions)
	log.Debugw("log request", "request", r)

	l, err := r.Stream(context.TODO())
	if err != nil {
		return
	}
	defer l.Close()

	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, l); err != nil {
		return
	}
	return buf.String(), nil
}

func GetServiceExternalIP(c *kubernetes.Clientset, log logging.Logger, name, ns string) (ip string, found bool, err error) {
	s, err := c.CoreV1().Services(ns).Get(context.TODO(), name, metav1.GetOptions{})
	log.Debugw("response for get service", "namespace", ns, "name", name, "response", s)
	if err != nil {
		return
	}

	switch len(s.Spec.ExternalIPs) {
	case 0:
		return
	default:
		found = true
		ip = s.Spec.ExternalIPs[0]
		return
	}
}

func GetSecretFieldDecoded(c *kubernetes.Clientset, log logging.Logger, secret, ns, field string) (string, error) {
	log.Debugw("getting secret", "namespace", ns, "secret", secret)
	s, err := c.CoreV1().Secrets(ns).Get(context.TODO(), secret, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	log.Debugw("got secret data", "namespace", ns, "secret", secret, "data", s.Data)

	fieldData, ok := s.Data[field]
	if !ok {
		return "", fmt.Errorf("%s field not found", field)
	}

	return string(fieldData), nil
}
