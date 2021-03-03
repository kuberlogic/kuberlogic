package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
)

func enforceJSONHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Incoming request received!")

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

func processAlertmanagerWebhook(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal("Error reading body!")
	}
	defer r.Body.Close()

	log.Printf("Raw body: %s\n", string(b))

	alert := &AlertWebhook{}
	if err := json.Unmarshal(b, &alert); err != nil {
		log.Fatal("Error unmarshalling Alertmanager webhook data!")
	}

	for _, a := range alert.Alerts {
		switch a.Status {
		case "resolved":
			err = a.resolve()
		case "firing":
			err = a.create()
		default:
			log.Fatalf("Unknown alert status: %s", alert.Status)
		}
	}

	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(fmt.Sprintf("Unexpected error during %s alert processing", alert.CommonLabels.Alertname)))
	}
}

func main() {
	initKubernetesClient()

	mux := http.NewServeMux()

	alertmanagerHandler := http.HandlerFunc(processAlertmanagerWebhook)
	mux.Handle("/", enforceJSONHandler(alertmanagerHandler))

	log.Println("Listening on :3000 port...")
	err := http.ListenAndServe(":3000", mux)
	log.Fatal(err)
}
