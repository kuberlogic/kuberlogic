/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package app

import (
	"fmt"
	subscriptionActions "github.com/chargebee/chargebee-go/actions/subscription"
	subscriptionModel "github.com/chargebee/chargebee-go/models/subscription"
)

const SubscriptionCreated = "subscription_created"

func setEndpoint(subscriptionId, endpoint string) error {
	_, err := subscriptionActions.UpdateForItems(subscriptionId, &subscriptionModel.UpdateForItemsRequestParams{}).
		AddParams("cf_domain", endpoint).Request()
	if err != nil {
		return err
	}
	return nil

}

func retriveSubscription(content map[string]interface{}) (*map[string]interface{}, error) {
	c, err := valueAsMap(content, "content")
	if err != nil {
		return nil, err
	}
	s, err := valueAsMap(*c, "subscription")
	if err != nil {
		return nil, err
	}
	return s, nil
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
