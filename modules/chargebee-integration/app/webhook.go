/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package app

import (
	"encoding/json"
	subscriptionModel "github.com/chargebee/chargebee-go/models/subscription"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"strings"
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
		if event["event_type"] != SubscriptionCreated {
			logger.Errorf("event type is unsupported: %s\n", event["event_type"])
			w.WriteHeader(http.StatusOK)
			return
		}

		subscription, err := GetSubscription(event)
		if err != nil {
			logger.Error("error retrieving subscription", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		logger = logger.With("subscription id", subscription.Id)
		logger.Infof("subscription status: %s", subscription.Status)

		svc := createServiceItem()
		svc.Subscription = subscription.Id

		err = ApplyMapping(logger, subscription, mapping, svc)
		if err != nil {
			logger.Error("error applying mapping", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = createService(logger, svc)
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
