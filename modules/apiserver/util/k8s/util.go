/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package k8s

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/config"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/logging"
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

func GetPodLogs(c kubernetes.Interface, log logging.Logger, name, container, ns string, lines int64) (logs string, err error) {
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

func GetServiceExternalIP(c kubernetes.Interface, log logging.Logger, name, ns string) (ip string, found bool, err error) {
	s, err := c.CoreV1().Services(ns).Get(context.TODO(), name, metav1.GetOptions{})
	log.Debugw("response for get service", "namespace", ns, "name", name, "response", s)
	if err != nil {
		return
	}

	if extName := s.Spec.ExternalName; extName != "" {
		found = true
		ip = extName
		return
	}

	if extIPs := s.Spec.ExternalIPs; len(extIPs) != 0 {
		found = true
		ip = extIPs[0]
		return
	}
	log.Debugw("service has no ExternalIPs. Checking LoadBalancers")

	if lbIPs := s.Status.LoadBalancer.Ingress; len(lbIPs) != 0 {
		found = true
		ip = lbIPs[0].IP
		return
	}
	return
}

func GetSecretFieldDecoded(c kubernetes.Interface, log logging.Logger, secret, ns, field string) (string, error) {
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
