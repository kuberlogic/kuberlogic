/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package main

import (
	"chargebee_integration/app"
	"fmt"
	"github.com/chargebee/chargebee-go"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	sentry2 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/sentry"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	rawLogger := initLogger()
	for _, value := range []string{
		"KUBERLOGIC_APISERVER_HOST",
		"KUBERLOGIC_APISERVER_SCHEME",
		"KUBERLOGIC_APISERVER_TOKEN",
		"CHARGEBEE_SITE",
		"CHARGEBEE_KEY",
		"KUBERLOGIC_DOMAIN",
		"KUBERLOGIC_TYPE",
		"SENTRY_DSN",
	} {
		if err := initEnv(value); err != nil {
			rawLogger.Warn(err.Error())
		} else {
			rawLogger.Debug("env", zap.String(value, viper.GetString(value)))
		}
	}

	// init sentry
	if dsn := viper.GetString("SENTRY_DSN"); dsn != "" {
		rawLogger = sentry2.UseSentryWithLogger(dsn, rawLogger, "chargebee-integration")

		err := sentry2.InitSentry(dsn, "chargebee-integration")
		if err != nil {
			rawLogger.Error(fmt.Sprintf("unable to init sentry: %v", err))
			os.Exit(1)
		}

		// Flush buffered events before the program terminates.
		defer sentry.Flush(2 * time.Second)

		rawLogger.Debug("sentry is initialized")
	}
	logger := rawLogger.Sugar()

	if viper.GetString("CHARGEBEE_SITE") != "" {
		chargebee.Configure(viper.GetString("CHARGEBEE_KEY"), viper.GetString("CHARGEBEE_SITE"))
		sentryHandler := sentryhttp.New(sentryhttp.Options{
			Repanic:         true,
			WaitForDelivery: true,
			Timeout:         5 * time.Second,
		})
		http.HandleFunc("/chargebee-webhook", sentryHandler.HandleFunc(app.WebhookHandler(logger)))
		logger.Info("webhook handler is initialized")
	} else {
		logger.Warn("ChargeBee site is not set. Requests will not be handled.")
	}

	addr := "0.0.0.0:4242"
	logger.Infof("Listening on %s\n", addr)
	logger.Fatal(http.ListenAndServe(addr, nil))
}

func initEnv(param string) error {
	_ = viper.BindEnv(param)
	value := viper.GetString(param)
	if value == "" {
		return fmt.Errorf("%s is not set", param)
	}
	return nil
}

func initLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.DisableStacktrace = true
	logger, _ := config.Build()
	defer func() {
		_ = logger.Sync()
	}()
	return logger
}
