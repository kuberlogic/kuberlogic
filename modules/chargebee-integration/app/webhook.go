/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package app

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	"go.uber.org/zap"
)

const (
	MaxSubscriptionRetries = 5
)

func WebhookHandler(baseLogger *zap.SugaredLogger, mapping []map[string]string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		defer func() {
			_ = req.Body.Close()
		}()

		event := make(map[string]interface{})
		err := decoder.Decode(&event)
		if err != nil {
			baseLogger.Error("request decode error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		logger := baseLogger.With("event id", event["id"])
		responseCode := http.StatusOK
		switch event["event_type"] {
		case SubscriptionCreated:
			responseCode = handleSubscriptionCreated(logger, event, mapping)
		case SubscriptionChangeed:
			responseCode = handleSubscriptionChanged(logger, event, mapping)
		case SubscriptionCancelled:
			go handleSubscriptionCancelled(logger, event)
		default:
			logger.Errorf("event type is unsupported: %s\n", event["event_type"])
		}
		w.WriteHeader(responseCode)
	}
}

func handleSubscriptionChanged(logger *zap.SugaredLogger, event map[string]interface{}, mapping []map[string]string) int { // int - http status
	subscription, err := GetSubscription(event)
	if err != nil {
		logger.Error("error retrieving subscription", err)
		return http.StatusBadRequest
	}

	logger = logger.With("subscription id", subscription.Id)
	// Check if service exists
	var existingService *models.Service
	for i := 0; i < 5; i++ {
		existingService, err = getServiceBySubscriptionId(logger, subscription.Id)
		if err != nil {
			logger.Error("Error getting service by subscription", err)
		} else {
			break
		}
		time.Sleep(time.Duration(i) * time.Second)
	}
	if err != nil {
		logger.Error("Retries exceeded while trying to get service by subscription", err)
		return http.StatusBadRequest
	}
	service := copyServiceItem(existingService) // avoid read-only fields

	entitlements, err := retrieveSubscriptionEntitlements(subscription.Id)
	if err != nil {
		logger.Error("error retrieving subscription entitlements: ", err)
		return http.StatusBadRequest
	}

	err = ApplyMapping(logger, entitlements, mapping, service)
	if err != nil {
		logger.Error("error applying mapping: ", err)
		return http.StatusBadRequest
	}

	err = editService(logger, service)
	if err != nil {
		logger.Error("edit operation error: ", err)
		return http.StatusServiceUnavailable
	}

	return http.StatusOK
}

func handleSubscriptionCreated(logger *zap.SugaredLogger, event map[string]interface{}, mapping []map[string]string) int { // int - http status
	subscription, err := GetSubscription(event)
	if err != nil {
		logger.Error("error retrieving subscription: ", err)
		return http.StatusBadRequest
	}

	logger = logger.With("subscription id", subscription.Id)
	logger.Infof("subscription status: %s", subscription.Status)

	svc := createServiceItem()
	svc.Subscription = subscription.Id

	entitlements, err := retrieveSubscriptionEntitlements(subscription.Id)
	if err != nil {
		logger.Error("error retrieving subscription entitlements: ", err)
		return http.StatusBadRequest
	}

	err = ApplyMapping(logger, entitlements, mapping, svc)
	if err != nil {
		logger.Error("error applying mapping: ", err)
		return http.StatusBadRequest
	}

	err = createService(logger, svc)
	if err != nil && checkAlreadyExists(err) {
		logger.Error("service already exists: ", err)
		// expected behavior due to prevent retries https://www.chargebee.com/docs/2.0/events_and_webhooks.html
		return http.StatusOK
	} else if err != nil {
		logger.Error("create operation error: ", err)
		return http.StatusServiceUnavailable
	}

	return http.StatusOK
}

func handleSubscriptionCancelled(logger *zap.SugaredLogger, event map[string]interface{}) {
	subscriptionId, err := GetSubscriptionId(event)
	if err != nil {
		logger.Error("Error extracting subscription from payload: ", err)
		return
	}
	logger = logger.With("subscription id", subscriptionId)
	// Check if service exists
	var service *models.Service
	for i := 0; i < MaxSubscriptionRetries; i++ {
		service, err = getServiceBySubscriptionId(logger, subscriptionId)
		if err != nil {
			logger.Error("Error getting service by subscription: ", err)
		} else {
			break
		}
		time.Sleep(time.Duration(i) * time.Second)
	}
	if err != nil {
		logger.Error("Retries exceeded while trying to get service by subscription: ", err)
		return
	}
	archiveService(logger, *service.ID)
}
