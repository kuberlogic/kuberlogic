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
	// init sentry
	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
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

	for _, value := range []string{
		"KUBERLOGIC_APISERVER_HOST",
		"KUBERLOGIC_APISERVER_SCHEME",
		"KUBERLOGIC_APISERVER_TOKEN",
		"CHARGEBEE_SITE",
		"CHARGEBEE_KEY",
		"KUBERLOGIC_DOMAIN",
		"KUBERLOGIC_TYPE",
	} {
		initEnv(logger, value)
	}

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

	addr := "localhost:4242"
	logger.Infof("Listening on %s\n", addr)
	logger.Fatal(http.ListenAndServe(addr, nil))
}

func initEnv(logger *zap.SugaredLogger, param string) {
	_ = viper.BindEnv(param)
	value := viper.GetString(param)
	if value == "" {
		logger.Warnf("parameter '%s' must be defined", param)
	}
	logger.Debugf("%s: %s", param, value)
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
