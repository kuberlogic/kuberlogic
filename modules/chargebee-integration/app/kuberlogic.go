/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package app

import (
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	client2 "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"strings"
	"time"
)

func createService(logger *zap.SugaredLogger, subscriptionId string) error {
	domain := viper.GetString("KUBERLOGIC_DOMAIN")
	serviceType := viper.GetString("KUBERLOGIC_TYPE")
	params := service.NewServiceAddParams()
	id := strings.ToLower(subscriptionId)
	params.ServiceItem = &models.Service{
		ID:     &id,
		Domain: domain,
		Type:   &serviceType,
	}

	apiClient, err := makeClient()
	if err != nil {
		return err
	}

	_, err = apiClient.Service.ServiceAdd(params,
		httptransport.APIKeyAuth("X-Token", "header", viper.GetString("KUBERLOGIC_APISERVER_TOKEN")))
	if err != nil {
		return err
	}

	logger.Infof("service is created: %s, waiting ready status", id)

	// checking the status asynchronously
	// maxRetries == 1 2 4 8  16 32 64 128 256 512 1024 2048 4096
	go checkStatus(logger, subscriptionId, time.Second, 13)
	return nil
}

func checkStatus(logger *zap.SugaredLogger, subscriptionId string, timeout time.Duration, maxRetries int) {
	if maxRetries <= 0 {
		return
	}

	apiClient, err := makeClient()
	if err != nil {
		logger.Error("create client error", err)
		return
	}

	id := strings.ToLower(subscriptionId)
	params := service.NewServiceGetParams()
	params.ServiceID = id
	response, err := apiClient.Service.ServiceGet(params,
		httptransport.APIKeyAuth("X-Token", "header", viper.GetString("KUBERLOGIC_APISERVER_TOKEN")))
	if err != nil {
		logger.Error("get operation error", err)
		return
	}

	if response.Payload != nil && response.Payload.Status == "ReadyConditionMet" {
		logger.Infof("status of service '%s' is ready", id)
		logger.Infof("endpoint is: %s", response.Payload.Endpoint)

		err = setEndpoint(subscriptionId, response.Payload.Endpoint)
		if err != nil {
			logger.Error("error updating subscription", err)
		}
		logger.Infof("Endpoint for subscription '%s' is updated", response.Payload.Endpoint)

		// TODO: send e-mail via chargebee
		return
	}

	newTimeout := timeout * 2
	newMaxRetries := maxRetries - 1
	logger.Infof("status is not ready, trying after %s. Left %d retries", newTimeout, newMaxRetries)
	time.Sleep(timeout)
	checkStatus(logger, subscriptionId, newTimeout, newMaxRetries)
}

func makeClient() (*client2.ServiceAPI, error) {
	hostname := viper.GetString("KUBERLOGIC_APISERVER_HOST")
	scheme := viper.GetString("KUBERLOGIC_APISERVER_SCHEME")

	r := httptransport.NewWithClient(hostname, client2.DefaultBasePath, []string{scheme}, nil)

	// set custom producer and consumer to use the default ones
	r.Consumers["application/json"] = runtime.JSONConsumer()
	r.Producers["application/json"] = runtime.JSONProducer()

	return client2.New(r, strfmt.Default), nil
}

func checkAlreadyExists(err error) bool {
	_, ok := err.(*service.ServiceAddConflict)
	return ok
}
