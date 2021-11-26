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

package main

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"time"
)

type CommonLabels struct {
	ResourceName string `json:"resourcename"`
	Alertname    string `json:"alertname"`
	Scope        string `json:"scope"`
	Namespace    string `json:"namespace"`
	Text         string `json:"text"`
	Severity     string `json:"severity"`
}

type Alert struct {
	Status string `json:"status"`
	Labels struct {
		CommonLabels
	} `json:"labels"`
	Annotations struct {
	} `json:"annotations"`
	StartsAt     time.Time `json:"startsAt"`
	EndsAt       time.Time `json:"endsAt"`
	GeneratorURL string    `json:"generatorURL"`
	Fingerprint  string    `json:"fingerprint"`
}

type AlertWebhookData struct {
	Receiver    string  `json:"receiver"`
	Status      string  `json:"status"`
	Alerts      []Alert `json:"alerts"`
	GroupLabels struct {
		Alertname string `json:"alertname"`
	} `json:"groupLabels"`
	CommonLabels struct {
		CommonLabels
	} `json:"commonLabels"`
	CommonAnnotations struct {
	} `json:"commonAnnotations"`
	ExternalURL     string `json:"externalURL"`
	Version         string `json:"version"`
	GroupKey        string `json:"groupKey"`
	TruncatedAlerts int    `json:"truncatedAlerts"`
}

type AlertStatus struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	GroupID    string `json:"group_id"`
	Expression string `json:"expression"`
	State      string `json:"state"`
	Value      string `json:"value"`
	Labels     struct {
	} `json:"labels"`
	Annotations struct {
	} `json:"annotations"`
	ActiveAt time.Time `json:"activeAt"`
}

const (
	instanceScope = "instance"
	serviceScope  = "service"
)

var (
	errUnknownScope = errors.New("unknown scope")
)

func (a Alert) alertname() string {
	return a.Labels.Alertname
}

func (a Alert) scope() string {
	return a.Labels.Scope
}

func (a Alert) targetResource() string {
	return a.Labels.ResourceName
}

func (a Alert) namespace() string {
	return a.Labels.Namespace
}

func (a Alert) fullName() string {
	return a.Fingerprint
}

func (a Alert) summary() string {
	return a.Labels.Text
}

func (a Alert) create(kubeRest rest.Interface, kubeClient kubernetes.Interface, log Logger, c *Config) error {
	status, err := a.getStatus()
	if err != nil {
		return errors.Wrap(err, "error getting status")
	}

	klsName, err := a.getServiceName(kubeClient, c.ServiceIdentifierLabel)
	if err != nil {
		return errors.Wrap(err, "error getting corresponding kuberlogicservicename")
	}
	podName := a.getPodName()

	if err := createAlertCR(a.fullName(), a.namespace(), a.alertname(), status.Value, klsName, podName, a.summary(), kubeRest); err != nil {
		log.Errorf("Error creating kuberlogic alert: %s", err)
		return errors.Wrap(err, "error creating kuberlogic alert")
	}
	log.Infof("Alert %s successfully created!", a.fullName())
	return nil
}

func (a Alert) resolve(kubeRest rest.Interface, log Logger) error {
	err := deleteAlertCR(a.fullName(), a.namespace(), kubeRest)
	if err != nil && !k8serrors.IsNotFound(err) {
		log.Errorf("Error resolving kuberlogic alert: %s", err)
		return errors.Wrap(err, "error deleting kuberlogicalert resource")
	}

	log.Infof("Alert %s successfully resolved!", a.fullName())
	return nil
}

func (a Alert) getServiceName(client kubernetes.Interface, serviceLabel string) (string, error) {
	if a.scope() == serviceScope {
		return a.targetResource(), nil
	}

	pod, err := client.CoreV1().Pods(a.namespace()).Get(context.Background(), a.targetResource(), v1.GetOptions{})
	if err != nil {
		return "", errors.Wrap(err, "error finding pod")
	}
	svcName := pod.Labels[serviceLabel]
	if svcName == "" {
		return "", errors.New("service label not found")
	}
	return svcName, nil
}

func (a Alert) getPodName() string {
	if a.scope() == serviceScope {
		return ""
	}
	return a.targetResource()
}

func (a Alert) getStatus() (*AlertStatus, error) {
	client := http.Client{
		Timeout: time.Second * 5,
	}
	r, err := client.Get(a.GeneratorURL)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching status")
	}

	s := &AlertStatus{}
	if err := json.NewDecoder(r.Body).Decode(s); err != nil {
		return nil, errors.Wrap(err, "error decoding status response")
	}
	return s, nil
}
