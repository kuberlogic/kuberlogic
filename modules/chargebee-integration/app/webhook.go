/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	subscriptionModel "github.com/chargebee/chargebee-go/models/subscription"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const ChargebeePrefixCustomField = "cf_"

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
		case SubscriptionCancelled:
			go handleSubscriptionCancelled(logger, event, mapping)
		default:
			logger.Errorf("event type is unsupported: %s\n", event["event_type"])
		}
		w.WriteHeader(responseCode)
	}
}

func handleSubscriptionCreated(logger *zap.SugaredLogger, event map[string]interface{}, mapping []map[string]string) int {
	subscription, err := GetSubscription(event)
	if err != nil {
		logger.Error("error retrieving subscription", err)
		return http.StatusBadRequest
	}

	logger = logger.With("subscription id", subscription.Id)
	logger.Infof("subscription status: %s", subscription.Status)

	svc := createServiceItem()
	svc.Subscription = subscription.Id

	err = ApplyMapping(logger, subscription, mapping, svc)
	if err != nil {
		logger.Error("error applying mapping", err)
		return http.StatusBadRequest
	}

	err = createService(logger, svc)
	if err != nil && checkAlreadyExists(err) {
		logger.Error("service already exists", err)
		// expected behavior due to prevent retries https://www.chargebee.com/docs/2.0/events_and_webhooks.html
		return http.StatusOK
	} else if err != nil {
		logger.Error("create operation error", err)
		return http.StatusServiceUnavailable
	}

	return http.StatusOK
}

func handleSubscriptionCancelled(logger *zap.SugaredLogger, event map[string]interface{}, mapping []map[string]string) {
	subscriptionId, err := GetSubscriptionId(event)
	if err != nil {
		logger.Error("Error extracting subscription from payload", err)
		return
	}
	logger = logger.With("subscription id", subscriptionId)
	// Check if service exists
	var service *models.Service
	for i := 0; i < 5; i++ {
		service, err = getServiceBySubscriptionId(logger, subscriptionId)
		if err != nil {
			logger.Error("Error getting service by subscription", err)
		} else {
			break
		}
		time.Sleep(time.Duration(i) * time.Second)
	}
	if err != nil {
		logger.Error("Retries exceeded while trying to get service by subscription", err)
		return
	}
	// Take backup of the service
	newBackup, err := addServiceBackup(logger, *service.ID)
	if err != nil {
		logger.Error("Error making service backup", err)
	} else {
		// Wait until backup is done
		err = waitForBackup(logger, *service.ID, newBackup.ID)
		// Delete all previous backups
		backups, err := listServiceBackups(logger, *service.ID)
		if err != nil {
			logger.Error("Error listing service backup", err)
			return
		}
		for _, backup := range backups {
			if backup.ID == newBackup.ID {
				continue
			}
			err = deleteServiceBackup(logger, backup.ID)
			if err != nil {
				logger.Error("Error deleting service backup", err)
				return
			}
		}
	}
	// Delete the application
	deleteService(logger, *service.ID)
}

func ApplyMapping(
	logger *zap.SugaredLogger,
	subscription *subscriptionModel.Subscription,
	mapping []map[string]string,
	svc *models.Service,
) error {
	for _, v := range subscription.SubscriptionItems {
		if v.ItemType == "plan" {
			logger.Debugw("subscription item is plan", "item price id", v.ItemPriceId)
			itemPrice, err := GetItemPrice(v.ItemPriceId)
			if err != nil {
				return errors.Wrapf(err, "item price is not retrived: %v\n", err)
			}

			for field, value := range itemPrice.CustomField {
				logger.Debugf("found custom field: %s = %v", field, value)
				if found := inMapping(mapping, field); found != "" {
					logger.Debugf("found cf in mapping: %s = %s", field, found)

					bytes, err := json.Marshal(unfold(found, value))
					if err != nil {
						return errors.Wrapf(err, "unable to encode value: %s\n", found)
					}

					err = json.Unmarshal(bytes, svc)
					if err != nil {
						return errors.Wrapf(err, "unable to decode value: %s\n", found)
					}
				} else {
					logger.Debug("not found in mapping ", field)
					if svc.Advanced == nil {
						svc.Advanced = make(map[string]interface{})
					}
					svc.Advanced[strings.TrimLeft(field, ChargebeePrefixCustomField)] = value
				}
			}
			break
		}
	}
	return nil
}

func inMapping(mapping []map[string]string, key string) string {
	for _, v := range mapping {
		if ChargebeePrefixCustomField+v["src"] == key {
			return v["dst"]
		}
	}
	return ""
}

func unfold(s string, value interface{}) map[string]interface{} {
	words := strings.Split(s, ".")

	var result, extendable map[string]interface{}
	for i, v := range words {
		if v != "" {
			if extendable == nil {
				extendable = make(map[string]interface{})
				result = extendable
			}
			if i == len(words)-1 {
				extendable[v] = value
			} else {
				extendable[v] = make(map[string]interface{})
				extendable = extendable[v].(map[string]interface{})
			}
		}
	}
	return result
}
