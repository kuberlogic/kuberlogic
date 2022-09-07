/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package main

import (
	"chargebee_integration/app"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/chargebee/chargebee-go"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/ghodss/yaml"
	sentry2 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/sentry"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	// version of package, substitute via ldflags
	ver string
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
		"KUBERLOGIC_TYPE",
		"SENTRY_DSN",
		"KUBERLOGIC_DEPLOYMENT_ID",
	} {
		if err := initEnv(value); err != nil {
			rawLogger.Warn(err.Error())
		} else {
			rawLogger.Debug("env", zap.String(value, viper.GetString(value)))
		}
	}
	deploymentId := viper.GetString("KUBERLOGIC_DEPLOYMENT_ID")

	// init sentry
	if dsn := viper.GetString("SENTRY_DSN"); dsn != "" {
		sentryTags := &sentry2.SentryTags{
			Component:    "chargebee-integration",
			Version:      ver,
			DeploymentId: deploymentId,
		}
		rawLogger = sentry2.UseSentryWithLogger(dsn, rawLogger, sentryTags)

		err := sentry2.InitSentry(dsn, sentryTags)
		if err != nil {
			rawLogger.Error(fmt.Sprintf("unable to init sentry: %v", err))
			os.Exit(1)
		}

		// Flush buffered events before the program terminates.
		defer sentry.Flush(2 * time.Second)

		rawLogger.Debug("sentry is initialized")
	}
	logger := rawLogger.Sugar()

	mapping, err := readMapping()
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	logger.Debug("mapping is configured: ", mapping)

	if viper.GetString("CHARGEBEE_SITE") != "" {
		chargebee.Configure(viper.GetString("CHARGEBEE_KEY"), viper.GetString("CHARGEBEE_SITE"))
		sentryHandler := sentryhttp.New(sentryhttp.Options{
			Repanic:         true,
			WaitForDelivery: true,
			Timeout:         5 * time.Second,
		})
		http.HandleFunc("/chargebee-webhook", sentryHandler.HandleFunc(app.WebhookHandler(logger, mapping)))
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

func readMapping() ([]map[string]string, error) {
	mappingFile := flag.String("mapping", "", "mapping file path")
	flag.Parse()

	mapping := make([]map[string]string, 0)
	if *mappingFile != "" {
		yamlFile, err := ioutil.ReadFile(*mappingFile)
		if err != nil {
			return nil, errors.Errorf("cannot read the file %s: #%v", *mappingFile, err)
		}
		err = yaml.Unmarshal(yamlFile, &mapping)
		if err != nil {
			return nil, errors.Errorf("cannot parse the file %s: #%v", *mappingFile, err)
		}
	}
	return mapping, nil
}
