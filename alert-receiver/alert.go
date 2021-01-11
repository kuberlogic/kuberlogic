package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Alert struct {
	Status string `json:"status"`
	Labels struct {
		Alertgroup          string `json:"alertgroup"`
		Alertname           string `json:"alertname"`
		ClusterType         string `json:"cluster_type"`
		ControlPlane        string `json:"control_plane"`
		Instance            string `json:"instance"`
		Job                 string `json:"job"`
		KubernetesNamespace string `json:"kubernetes_namespace"`
		KubernetesPodName   string `json:"kubernetes_pod_name"`
		Name                string `json:"name"`
		Namespace           string `json:"namespace"`
		PodTemplateHash     string `json:"pod_template_hash"`
	} `json:"labels"`
	Annotations struct {
	} `json:"annotations"`
	StartsAt     time.Time `json:"startsAt"`
	EndsAt       time.Time `json:"endsAt"`
	GeneratorURL string    `json:"generatorURL"`
	Fingerprint  string    `json:"fingerprint"`
}

type AlertWebhook struct {
	Receiver    string  `json:"receiver"`
	Status      string  `json:"status"`
	Alerts      []Alert `json:"alerts"`
	GroupLabels struct {
		Alertname string `json:"alertname"`
	} `json:"groupLabels"`
	CommonLabels struct {
		Alertgroup          string `json:"alertgroup"`
		Alertname           string `json:"alertname"`
		ClusterType         string `json:"cluster_type"`
		ControlPlane        string `json:"control_plane"`
		Instance            string `json:"instance"`
		Job                 string `json:"job"`
		KubernetesNamespace string `json:"kubernetes_namespace"`
		KubernetesPodName   string `json:"kubernetes_pod_name"`
		Name                string `json:"name"`
		Namespace           string `json:"namespace"`
		PodTemplateHash     string `json:"pod_template_hash"`
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
		Alertgroup          string `json:"alertgroup"`
		ClusterType         string `json:"cluster_type"`
		ControlPlane        string `json:"control_plane"`
		Instance            string `json:"instance"`
		Job                 string `json:"job"`
		KubernetesNamespace string `json:"kubernetes_namespace"`
		KubernetesPodName   string `json:"kubernetes_pod_name"`
		Name                string `json:"name"`
		Namespace           string `json:"namespace"`
		PodTemplateHash     string `json:"pod_template_hash"`
	} `json:"labels"`
	Annotations struct {
	} `json:"annotations"`
	ActiveAt time.Time `json:"activeAt"`
}

func (a Alert) getName() string {
	return fmt.Sprintf("%s-%s-%s", a.Labels.Name, a.Labels.Alertname)
}

func (a Alert) create() error {
	name := a.getName()
	status := &AlertStatus{}

	if err := a.getStatus(status); err != nil {
		log.Printf("Couldn't fetch alert status")
		return err
	}

	if err := createAlertCR(name, "default", a.Labels.Alertname, status.Value, a.Labels.Name, status.Labels.KubernetesPodName); err != nil {
		log.Printf("Error creating cloudmanaged alert: %s", err)
		return err
	}
	log.Printf("Alert %s succesfully created!", name)
	return nil
}

func (a Alert) resolve() error {
	name := a.getName()

	err := deleteAlertCR(name, a.Labels.Namespace)
	if err != nil {
		log.Printf("Error resolving cloudmanaged alert: %s", err)
		return err
	}

	log.Printf("Alert %s succesfully resolved!", name)
	return nil
}

func (a Alert) getStatus(s *AlertStatus) error {
	r, err := http.Get(a.GeneratorURL)
	if err != nil {
		log.Printf("Error fetching status for %s", a.Labels.Name)
		return err
	}

	if err := json.NewDecoder(r.Body).Decode(s); err != nil {
		log.Printf("Error unmarshalling status for %s: %s", a.Labels.Name)
		return err
	}
	log.Printf("Succcesfuly got status for %s", a.Labels.Name)
	return nil
}
