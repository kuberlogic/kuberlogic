/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package app

import (
	subscriptionModel "github.com/chargebee/chargebee-go/models/subscription"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	"go.uber.org/zap"
	"reflect"
	"testing"
)

func TestUnfold(t *testing.T) {
	actual := unfold("A.B.C.D", 100)
	expected := map[string]interface{}{
		"A": map[string]interface{}{
			"B": map[string]interface{}{
				"C": map[string]interface{}{
					"D": 100,
				},
			},
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("actual vs expected: %v vs %v", actual, expected)
	}

	actual = unfold("A.B.", 100)
	expected = map[string]interface{}{
		"A": map[string]interface{}{
			"B": map[string]interface{}{},
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("actual vs expected: %v vs %v", actual, expected)
	}

	actual = unfold(".A.B", "A")
	expected = map[string]interface{}{
		"A": map[string]interface{}{
			"B": "A",
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("actual vs expected: %v vs %v", actual, expected)
	}

	actual = unfold("A.B;C.D", 100)
	expected = map[string]interface{}{
		"A": map[string]interface{}{
			"B;C": map[string]interface{}{
				"D": 100,
			},
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("actual vs expected: %v vs %v", actual, expected)
	}
}

func TestInMapping(t *testing.T) {
	actual := inMapping([]map[string]string{{
		"src": "custom-field",
		"dst": "result",
	}}, "custom-field")
	expected := ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("actual vs expected: %v vs %v", actual, expected)
	}

	actual = inMapping([]map[string]string{{
		"src": "custom-field",
		"dst": "result",
	}}, ChargebeePrefixCustomField+"custom-field")
	expected = "result"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("actual vs expected: %v vs %v", actual, expected)
	}
}

func TestApplyMappingEmptySubscriptionItems(t *testing.T) {
	baseLogger, _ := zap.NewDevelopmentConfig().Build()
	err := ApplyMapping(
		baseLogger.Sugar(),
		&subscriptionModel.Subscription{},
		[]map[string]string{{
			"src": "volumeSize",
			"dst": "limits.volumeSize",
		}},
		&models.Service{},
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
