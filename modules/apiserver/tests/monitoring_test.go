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

package tests

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/common/log"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

type tMonitoring struct {
	// vmEndpoint points to a deployed and configured VictoriaMetrics
	vmEndpoint string

	// test SMTP server
	mailcatcherEndpoint string

	// kuberlogicservices that are expected to be present and monitored
	mysqlService tService
	pgService    tService
}

func (tm tMonitoring) CheckTargets() func(t *testing.T) {
	return func(t *testing.T) {
		res, err := http.Get(tm.vmEndpoint + "/api/v1/targets")
		if err != nil || res.StatusCode != http.StatusOK {
			t.Fatalf("error getting victoriametrics targets: %v status code %d", err, res.StatusCode)
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("error reading victoriametrics response: %v", err)
		}

		type vmTarget struct {
			DiscoveredLabels   map[string]string `json:"discoveredLabels,omitempty"`
			Labels             map[string]string `json:"labels,omitempty"`
			ScrapePool         string            `json:"scrapePool,omitempty"`
			ScrapeUrl          string            `json:"scrapeUrl,omitempty"`
			LastError          string            `json:"lastError,omitempty"`
			LastScrape         string            `json:"lastScrape,omitempty"`
			LastScrapeDuration float64           `json:"lastScrapeDuration,omitempty"`
			LastSamplesScraped int               `json:"lastSamplesScraped,omitempty"`
			Health             string            `json:"health,omitempty"`
		}

		type vmTargets struct {
			Status string `json:"status"`
			Data   *struct {
				ActiveTargets  []vmTarget  `json:"activeTargets"`
				DroppedTargets []vmTargets `json:"droppedTargets,omitempty"`
			} `json:"data,omitempty"`
		}

		data := &vmTargets{}
		if err := json.Unmarshal(body, data); err != nil {
			t.Fatalf("error decoding victoriametrics response: %v", err)
		}

		var (
			// nodeJob is a job for Kubernetes node monitoring
			nodeJob = "kubernetes-nodes"
			// podJob is a job for Kubernetes pods monitoring
			podJob = "kubernetes-pods"
			// scrapePool is a job for static services monitoring
			scrapePool = "kubernetes-stats-services"
			// kube-eagle monitoring instance
			kubeEagle = "kube-eagle:8443"
			// kube-state-metrics monitoring instance
			kubeStateMetrics = "kube-state-metrics:8443"
			// mysql first pod
			mysql = tm.mysqlService.name + "-mysql-0"
			// postgres first pod
			pg = "kuberlogic-" + tm.pgService.name
		)

		expectedActiveTargets := map[string]bool{
			// at least one node
			nodeJob: false,

			kubeEagle:        false,
			kubeStateMetrics: false,

			// at least one pod for kuberlogicservice
			mysql: false,
			pg:    false,
		}

		for _, t := range data.Data.ActiveTargets {
			// skip down or errored targets
			if t.Health != "up" || t.LastError != "" {
				continue
			}

			switch t.ScrapePool {
			case nodeJob:
				expectedActiveTargets[nodeJob] = true
			case podJob:
				switch t.Labels["kubernetes_pod_name"] {
				case mysql:
					expectedActiveTargets[mysql] = true
				case pg:
					expectedActiveTargets[pg] = true
				}
			case scrapePool:
				switch t.Labels["instance"] {
				case kubeEagle:
					expectedActiveTargets[kubeEagle] = true
				case kubeStateMetrics:
					expectedActiveTargets[kubeStateMetrics] = true
				}
			}
		}

		for target, active := range expectedActiveTargets {
			if !active {
				t.Errorf("target %s is not active", target)
			}
		}
	}
}

func (tm *tMonitoring) CheckAlertNotification() func(t *testing.T) {
	return func(t *testing.T) {
		// create a restore job that will fail
		api := newApi(t)
		api.setBearerToken()
		api.setRequestBody(`     {
        "key": "s3://nonexistend",
        "database": "nonexistent"
     }`)
		api.sendRequestTo(http.MethodPost, fmt.Sprintf("/services/%s:%s/restores", tm.pgService.ns, tm.pgService.name))
		api.responseCodeShouldBe(200)

		// poll mailcatcher for the failed restore alert
		type MailcatcherMessage struct {
			Sender     string   `json:"sender"`
			Recipients []string `json:"recipients"`
			Subject    string   `json:"subject"`
		}
		messages := make([]*MailcatcherMessage, 0)
		var found bool

		for i := 0; i < 300; i += 1 {
			if found {
				log.Infof("Alert notification received")
				break
			}
			time.Sleep(time.Duration(1) * time.Second)
			r, err := http.Get(tm.mailcatcherEndpoint + "/messages")
			if err != nil {
				t.Errorf("error getting mailcatcher messages: %v", err)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("error reading mailcatcher response data: %v", err)
			}
			r.Body.Close()
			if err := json.Unmarshal(body, &messages); err != nil {
				t.Errorf("error decoding mailcatcher response data: %v", err)
			}
			for _, m := range messages {
				if fmt.Sprintf("CRITICAL: SERVICE %s ALERT kuberlogicrestore-failed", tm.pgService.name) == m.Subject {
					found = true
				}
			}
			log.Infof("Still no alert notification after: %ds", i)
		}
	}
}

func TestMonitoringStack(t *testing.T) {
	tm := &tMonitoring{
		vmEndpoint:          os.Getenv("VICTORIAMETRICS_ENDPOINT"),
		mailcatcherEndpoint: os.Getenv("MAILCATCHER_ENDPOINT"),
		mysqlService:        mysqlTestService,
		pgService:           pgTestService,
	}

	t.Run("victoriaMetrics active targets", tm.CheckTargets())
	t.Run("alerts notifications", tm.CheckAlertNotification())
}
