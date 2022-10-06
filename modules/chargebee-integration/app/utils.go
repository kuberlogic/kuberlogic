/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package app

import (
	"encoding/json"
	subscriptionEntitlementModel "github.com/chargebee/chargebee-go/models/subscriptionentitlement"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strings"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
)

func ApplyMapping(
	logger *zap.SugaredLogger,
	entitlements []*subscriptionEntitlementModel.SubscriptionEntitlement,
	mapping []map[string]string,
	svc *models.Service,
) error {
	// now supports only string fields as "src" in mapping
	for _, elem := range entitlements {
		logger.Debugf("found feature: %s = %v", elem.FeatureId, elem.Value)
		if found := inMapping(mapping, elem.FeatureId); found != "" {
			logger.Debugf("found feature in mapping: %s = %s", elem.FeatureId, found)

			bytes, err := json.Marshal(unfold(found, elem.Value))
			if err != nil {
				return errors.Wrapf(err, "unable to encode value: %s\n", found)
			}

			err = json.Unmarshal(bytes, svc)
			if err != nil {
				return errors.Wrapf(err, "unable to decode value: %s\n", found)
			}
		} else {
			logger.Debug("not found in mapping ", elem.FeatureId)
			if svc.Advanced == nil {
				svc.Advanced = make(map[string]interface{})
			}
			svc.Advanced[elem.FeatureId] = elem.Value
		}
	}
	return nil
}

func inMapping(mapping []map[string]string, key string) string {
	for _, v := range mapping {
		if v["src"] == key {
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
