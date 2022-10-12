/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package app

import (
	"fmt"

	itemPriceActions "github.com/chargebee/chargebee-go/actions/itemprice"
	subscriptionActions "github.com/chargebee/chargebee-go/actions/subscription"
	subscriptionEntitlementAction "github.com/chargebee/chargebee-go/actions/subscriptionentitlement"
	itemPriceModel "github.com/chargebee/chargebee-go/models/itemprice"
	subscriptionModel "github.com/chargebee/chargebee-go/models/subscription"
	subscriptionEntitlementModel "github.com/chargebee/chargebee-go/models/subscriptionentitlement"
	"github.com/pkg/errors"
)

const (
	SubscriptionCreated   = "subscription_created"
	SubscriptionChangeed  = "subscription_changed"
	SubscriptionCancelled = "subscription_cancelled"
)

func setEndpoint(subscriptionId, endpoint string) error {
	_, err := subscriptionActions.UpdateForItems(subscriptionId, &subscriptionModel.UpdateForItemsRequestParams{}).
		AddParams("cf_domain", endpoint).Request()
	if err != nil {
		return err
	}
	return nil

}

func GetSubscriptionId(content map[string]interface{}) (string, error) {
	c, err := valueAsMap(content, "content")
	if err != nil {
		return "", errors.Wrap(err, "content section does not exist")
	}
	s, err := valueAsMap(*c, "subscription")
	if err != nil {
		return "", errors.Wrap(err, "subscription section does not exist")
	}

	id, ok := (*s)["id"].(string)
	if !ok {
		return "", errors.Wrapf(err, "subscription is not type string: %v\n", id)
	}
	return id, nil
}

func GetSubscription(content map[string]interface{}) (*subscriptionModel.Subscription, error) {
	id, err := GetSubscriptionId(content)
	if err != nil {
		return nil, err
	}
	subscription, err := retrieveSubscription(id)
	if err != nil {
		return nil, errors.Wrapf(err, "subscription is not retrived with id: %s", id)
	}
	return subscription, nil
}

func valueAsMap(content map[string]interface{}, value string) (*map[string]interface{}, error) {
	if v, ok := content[value]; ok {
		if result, ok := v.(map[string]interface{}); ok {
			return &result, nil
		}
		return nil, fmt.Errorf("%s does not converted correctly: %v", value, content)
	}
	return nil, fmt.Errorf("%s section does not exist: %v", value, content)
}

func retrieveSubscription(id string) (*subscriptionModel.Subscription, error) {
	result, err := subscriptionActions.Retrieve(id).Request()
	if err != nil {
		return nil, err
	}
	return result.Subscription, nil
}

func GetItemPrice(id string) (*itemPriceModel.ItemPrice, error) {
	result, err := itemPriceActions.Retrieve(id).Request()
	if err != nil {
		return nil, err
	}
	return result.ItemPrice, nil
}

func retrieveSubscriptionEntitlements(id string) ([]*subscriptionEntitlementModel.SubscriptionEntitlement, error) {
	includeDrafts := false
	params := &subscriptionEntitlementModel.SubscriptionEntitlementsForSubscriptionRequestParams{
		IncludeDrafts: &includeDrafts,
	}
	result, err := subscriptionEntitlementAction.SubscriptionEntitlementsForSubscription(id, params).ListRequest()
	if err != nil {
		return nil, err
	}
	entitlements := make([]*subscriptionEntitlementModel.SubscriptionEntitlement, 0)
	for _, item := range result.List {
		entitlements = append(entitlements, item.SubscriptionEntitlement)
	}
	return entitlements, nil
}
