package cfg

import (
	"github.com/vrischmann/envconfig"
)

type Grafana struct {
	Enabled  bool   `envconfig:"default=false,optional"`
	Endpoint string `envconfig:"optional"`
	Login    string `envconfig:"default=admin,optional"`
	Password string `envconfig:"default=admin,optional"`
}

type Config struct {
	MetricsAddr          string `envconfig:"default=:8080,optional"`
	EnableLeaderElection bool   `envconfig:"default=false,optional"`

	ImageRepo           string `envconfig:"IMG_REPO"`
	ImagePullSecretName string `envconfig:"IMG_PULL_SECRET"`
	Namespace           string `envconfig:"POD_NAMESPACE"`

	SentryDsn string `envconfig:"optional"`

	NotificationChannels struct {
		EmailEnabled bool                           `json:"default=false,optional"`
		Email        EmailNotificationChannelConfig `json:"optional"`
	} `envconfig:"optional"`

	Grafana Grafana `envconfig:"optional"`
}

type EmailNotificationChannelConfig struct {
	Host string `envconfig:"optional"`
	Port int    `envconfig:"optional"`
	TLS  struct {
		Insecure bool `envconfig:"optional"`
		Enabled  bool `envconfig:"optional"`
	} `envconfig:"optional"`
	Username string `envconfig:"optional"`
	Password string `envconfig:"optional"`
}

func NewConfig() (*Config, error) {
	cfg := new(Config)

	if err := envconfig.Init(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
