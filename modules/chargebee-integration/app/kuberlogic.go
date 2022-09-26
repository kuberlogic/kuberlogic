/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package app

import (
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	client2 "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/backup"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func createService(logger *zap.SugaredLogger, svc *models.Service) error {
	params := service.NewServiceAddParams()
	params.ServiceItem = svc

	apiClient, err := makeClient()
	if err != nil {
		return err
	}

	_, err = apiClient.Service.ServiceAdd(params,
		httptransport.APIKeyAuth("X-Token", "header", viper.GetString("KUBERLOGIC_APISERVER_TOKEN")))
	if err != nil {
		return err
	}

	logger.Infof("service is created: %s, waiting ready status", svc.ID)

	// checking the status asynchronously
	// maxRetries == 1 2 4 8  16 32 64 128 256 512 1024 2048 4096
	go checkStatus(logger, *svc.ID, svc.Subscription, time.Second, 13)
	return nil
}

func deleteService(logger *zap.SugaredLogger, serviceId string) error {
	params := service.NewServiceDeleteParams()
	params.ServiceID = serviceId

	apiClient, err := makeClient()
	if err != nil {
		return err
	}
	_, err = apiClient.Service.ServiceDelete(
		params,
		httptransport.APIKeyAuth("X-Token", "header", viper.GetString("KUBERLOGIC_APISERVER_TOKEN")),
	)
	return nil
}

func getServiceBySubscriptionId(logger *zap.SugaredLogger, subscriptionId string) (*models.Service, error) {
	params := service.NewServiceListParams()
	params.SubscriptionID = &subscriptionId

	apiClient, err := makeClient()
	if err != nil {
		return nil, err
	}

	response, err := apiClient.Service.ServiceList(
		params,
		httptransport.APIKeyAuth("X-Token", "header", viper.GetString("KUBERLOGIC_APISERVER_TOKEN")),
	)
	if err != nil {
		return nil, err
	}
	if len(response.Payload) > 0 {
		return response.Payload[0], nil
	}
	return nil, errors.Errorf("There is no service with subscription id: %s", subscriptionId)
}

func addServiceBackup(logger *zap.SugaredLogger, serviceID string) (*models.Backup, error) {
	params := backup.NewBackupAddParams()
	params.BackupItem = new(models.Backup)
	params.BackupItem.ServiceID = serviceID

	apiClient, err := makeClient()
	if err != nil {
		return nil, err
	}
	response, err := apiClient.Backup.BackupAdd(
		params,
		httptransport.APIKeyAuth("X-Token", "header", viper.GetString("KUBERLOGIC_APISERVER_TOKEN")),
	)
	if err != nil {
		return nil, err
	}
	return response.Payload, nil
}

func listServiceBackups(logger *zap.SugaredLogger, serviceID string) (models.Backups, error) {
	params := backup.NewBackupListParams()
	params.ServiceID = &serviceID

	apiClient, err := makeClient()
	if err != nil {
		return nil, err
	}
	response, err := apiClient.Backup.BackupList(
		params,
		httptransport.APIKeyAuth("X-Token", "header", viper.GetString("KUBERLOGIC_APISERVER_TOKEN")),
	)
	if err != nil {
		return nil, err
	}
	return response.Payload, nil
}

func deleteServiceBackup(logger *zap.SugaredLogger, backupID string) error {
	params := backup.NewBackupDeleteParams()
	params.BackupID = backupID

	apiClient, err := makeClient()
	if err != nil {
		return err
	}
	_, err = apiClient.Backup.BackupDelete(
		params,
		httptransport.APIKeyAuth("X-Token", "header", viper.GetString("KUBERLOGIC_APISERVER_TOKEN")),
	)
	if err != nil {
		return err
	}
	return nil
}

func waitForBackup(logger *zap.SugaredLogger, serviceID string, backupID string) error {
	timeout := time.Second
	for i := 13; i > 0; i-- {
		backups, err := listServiceBackups(logger, serviceID)
		if err != nil {
			return err
		}
		for _, backup := range backups {
			if backup.ID == backupID {
				switch backup.Status {
				case v1alpha1.KlbSuccessfulCondType:
					return nil
				case v1alpha1.KlbFailedCondType:
					return errors.New("Error occured while creating service backup")
				}
			}
		}
		timeout := timeout * 2
		time.Sleep(timeout)
		logger.Infof("backup is not ready, trying after %s. Left %d retries", timeout, i-1)
	}
	return errors.New("Retries exceeded, backup is not ready")
}

func checkStatus(logger *zap.SugaredLogger, name, subscriptionId string, timeout time.Duration, maxRetries int) {
	newTimeout := timeout
	if maxRetries <= 0 {
		return
	}

	apiClient, err := makeClient()
	if err != nil {
		logger.Error("create client error", err)
		return
	}

	params := service.NewServiceGetParams()
	params.ServiceID = name
	response, err := apiClient.Service.ServiceGet(params,
		httptransport.APIKeyAuth("X-Token", "header", viper.GetString("KUBERLOGIC_APISERVER_TOKEN")))
	if err != nil {
		logger.Error("get operation error", err)
		return
	}

	if response.Payload != nil && response.Payload.Status == v1alpha1.ReadyCondType {
		logger.Infof("status of service '%s' is ready", name)
		logger.Infof("endpoint is: %s", response.Payload.Endpoint)

		err = setEndpoint(subscriptionId, response.Payload.Endpoint)
		if err != nil {
			logger.Error("error updating subscription", err)
		}
		logger.Infof("Endpoint for subscription '%s' is updated", response.Payload.Endpoint)

		return
	}

	newTimeout = newTimeout * 2
	newMaxRetries := maxRetries - 1
	logger.Infof("status is not ready, trying after %s. Left %d retries", newTimeout, newMaxRetries)
	time.Sleep(timeout)
	checkStatus(logger, name, subscriptionId, newTimeout, newMaxRetries)
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

func createServiceItem() *models.Service {
	name := petname.Generate(2, "-")
	serviceType := viper.GetString("KUBERLOGIC_TYPE")
	return &models.Service{
		ID:   &name,
		Type: &serviceType,
	}
}
