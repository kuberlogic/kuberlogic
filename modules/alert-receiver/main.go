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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"mime"
	"net/http"
)

const (
	alertResolvedState = "resolved"
	alertFiringState   = "firing"
)

func enforceJSONHandler(next http.Handler, log Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("New alert processing request")
		contentType := r.Header.Get("Content-Type")

		if contentType != "" {
			mt, _, err := mime.ParseMediaType(contentType)
			if err != nil {
				http.Error(w, "Malformed Content-Type header", http.StatusBadRequest)
				return
			}

			if mt != "application/json" {
				http.Error(w, "Content-Type header must be application/json", http.StatusUnsupportedMediaType)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func setupAlertsProcessor(kubeClientSet kubernetes.Interface, kubeRestClient rest.Interface, log Logger, cfg *Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b, processingErr := ioutil.ReadAll(r.Body)
		if processingErr != nil {
			log.Errorf("Error reading body!")
		}
		defer r.Body.Close()

		log.Debugf("Raw body: %s\n", string(b))

		alert := &AlertWebhookData{}
		if processingErr = json.Unmarshal(b, &alert); processingErr != nil {
			log.Errorf("Error unmarshalling Alertmanager webhook data with body: `%v`", string(b))
		}

		for _, a := range alert.Alerts {
			switch a.Status {
			case alertResolvedState:
				processingErr = a.resolve(kubeRestClient, log)
			case alertFiringState:
				processingErr = a.create(kubeRestClient, kubeClientSet, log, cfg)
			default:
				log.Fatalf("Unknown alert status: %s", alert.Status)
			}
		}

		if processingErr != nil {
			log.Errorf("Unexpected error `%v` during alert `%s` processing", processingErr, alert.CommonLabels.Alertname)
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(fmt.Sprintf("Unexpected error during %s alert processing", alert.CommonLabels.Alertname)))
		}
	}
}

func main() {
	log := newLogger(true)
	kubeClientSet, kubeRestClient, err := newKubernetesClients()
	if err != nil {
		log.Fatalf("Error creating Kubernetes clients: %v", err)
	}
	cfg, err := newConfig()
	if err != nil {
		log.Fatalf("Error creating config: %v", err)
	}
	log = newLogger(cfg.DebugLogs)

	mux := http.NewServeMux()

	alertmanagerHandler := http.HandlerFunc(setupAlertsProcessor(kubeClientSet, kubeRestClient, log, cfg))
	mux.Handle("/", enforceJSONHandler(alertmanagerHandler, log))

	log.Infof("Listening on :%d port", cfg.Port)
	if err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), mux); err != nil {
		log.Fatalf("Fatal error during startup: %v", err)
	}
}
