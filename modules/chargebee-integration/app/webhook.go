/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package app

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

func WebhookHandler(baseLogger *zap.SugaredLogger) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		defer func() {
			_ = req.Body.Close()
		}()

		event := make(map[string]interface{})
		//_ := new(Event)
		err := decoder.Decode(&event)
		if err != nil {
			baseLogger.Error("request decode error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		logger := baseLogger.With("event id", event["id"])
		if event["event_type"] != SubscriptionCreated {
			logger.Errorf("event type is unsupported: %s\n", event["event_type"])
			w.WriteHeader(http.StatusOK)
			return
		}

		subscription, err := retriveSubscription(event)
		if err != nil {
			logger.Error("error retrieving subscription", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id, ok := (*subscription)["id"].(string)
		if !ok {
			logger.Errorf("subscription is not type string: %v\n", id)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logger = logger.With("subscription id", id)
		logger.Infof("subscription status: %s\n", (*subscription)["status"])

		err = createService(logger, id)
		if err != nil && checkAlreadyExists(err) {
			logger.Error("service already exists", err)
			// expected behavior due to prevent retries https://www.chargebee.com/docs/2.0/events_and_webhooks.html
			w.WriteHeader(http.StatusOK)
			return
		} else if err != nil {
			logger.Error("create operation error", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
