/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package main

import (
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
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/kuberlogic/kuberlogic/modules/chargebee-integration/app"
	"github.com/kuberlogic/kuberlogic/modules/chargebee-integration/cfg"
	sentry2 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/sentry"
)

var (
	// version of package, substitute via ldflags
	ver string
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	rawLogger := initLogger()

	for _, value := range []string{
		cfg.KlApiserverHostParam,
		cfg.KlApiserverSchemeParam,
		cfg.KlApiserverTokenParam,
		cfg.ChargebeeSiteParam,
		cfg.ChargebeeKeyParam,
		cfg.KlTypeParam,
		cfg.SentryDsnParam,
		cfg.KlDeploymentIDParam,
		cfg.AuthUserParam,
		cfg.AuthPasswordParam,
	} {
		if err := cfg.InitEnv(value); err != nil {
			rawLogger.Warn(err.Error())
		} else {
			rawLogger.Debug("env", zap.String(value, viper.GetString(value)))
		}
	}
	deploymentId := viper.GetString(cfg.KlDeploymentIDParam)

	// init sentry
	if dsn := viper.GetString(cfg.SentryDsnParam); dsn != "" {
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

	if viper.GetString(cfg.ChargebeeSiteParam) != "" {
		chargebee.Configure(viper.GetString(cfg.ChargebeeKeyParam), viper.GetString(cfg.ChargebeeSiteParam))
		sentryHandler := sentryhttp.New(sentryhttp.Options{
			Repanic:         true,
			WaitForDelivery: true,
			Timeout:         5 * time.Second,
		})
		http.HandleFunc("/chargebee-webhook", newAuthenticationHandler(logger)(sentryHandler.HandleFunc(app.WebhookHandler(logger, mapping))))
		logger.Info("webhook handler is initialized")
	} else {
		logger.Warn("ChargeBee site is not set. Requests will not be handled.")
	}

	addr := "0.0.0.0:4242"
	logger.Infof("Listening on %s\n", addr)
	logger.Fatal(http.ListenAndServe(addr, nil))
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

func newAuthenticationHandler(log *zap.SugaredLogger) func(h http.HandlerFunc) http.HandlerFunc {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()

			if !ok || username != viper.GetString(cfg.AuthUserParam) || password != viper.GetString(cfg.AuthPasswordParam) {
				log.Warn("Authentication failed", username, ":", password, " ", viper.GetString(cfg.AuthUserParam), ":", viper.GetString(cfg.AuthPasswordParam))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			}
		}
	}
}
