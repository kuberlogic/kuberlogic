/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package main

import (
	"chargebee_integration/app"
	"github.com/chargebee/chargebee-go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	logger := initLogger()

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

	chargebee.Configure(viper.GetString("CHARGEBEE_KEY"), viper.GetString("CHARGEBEE_SITE"))

	http.HandleFunc("/chanrgebee-webhook", app.WebhookHandler(logger))
	addr := "localhost:4242"
	logger.Infof("Listening on %s\n", addr)
	logger.Fatal(http.ListenAndServe(addr, nil))
}

func initEnv(logger *zap.SugaredLogger, param string) {
	_ = viper.BindEnv(param)
	value := viper.GetString(param)
	if value == "" {
		logger.Fatalf("parameter '%s' must be defined", param)
	}
	logger.Debugf("%s: %s", param, value)
}

func initLogger() *zap.SugaredLogger {
	config := zap.NewDevelopmentConfig()
	config.DisableStacktrace = true
	logger, _ := config.Build()
	defer func() {
		_ = logger.Sync()
	}()
	sugar := logger.Sugar()
	return sugar
}
