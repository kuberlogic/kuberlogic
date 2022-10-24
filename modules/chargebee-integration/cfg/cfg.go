package cfg

import (
	"fmt"

	"github.com/spf13/viper"
)

const (
	SentryDsnParam         = "SENTRY_DSN"
	KlApiserverHostParam   = "KUBERLOGIC_APISERVER_HOST"
	KlApiserverSchemeParam = "KUBERLOGIC_APISERVER_SCHEME"
	KlApiserverTokenParam  = "KUBERLOGIC_APISERVER_TOKEN"
	ChargebeeSiteParam     = "CHARGEBEE_SITE"
	ChargebeeKeyParam      = "CHARGEBEE_KEY"
	KlTypeParam            = "KUBERLOGIC_TYPE"
	KlDeploymentIDParam    = "KUBERLOGIC_DEPLOYMENT_ID"
	AuthUserParam          = "CHARGEBEE_WEBHOOK_USER"
	AuthPasswordParam      = "CHARGEBEE_WEBHOOK_PASSWORD"
)

func InitEnv(param string) error {
	_ = viper.BindEnv(param)
	value := viper.GetString(param)
	if value == "" {
		return fmt.Errorf("%s is not set", param)
	}
	return nil
}
