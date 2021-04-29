package cfg

import "github.com/vrischmann/envconfig"

type Config struct {
	MetricsAddr          string `envconfig:"default=:8080,optional"`
	EnableLeaderElection bool   `envconfig:"default=false,optional"`

	ImageRepo           string `envconfig:"IMG_REPO"`
	ImagePullSecretName string `envconfig:"IMG_PULL_SECRET"`

	SentryDsn string `envconfig:"optional"`

	NotificationChannels struct {
		Email struct {
			Host string
			Port string
			TLS  struct {
				Insecure bool
				Enabled  bool
			} `envconfig:"optional"`
			Username string `envconfig:"optional"`
			Password string `envconfig:"optional"`
		} `envconfig:"optional"`
	} `envconfig:"optional"`
}

func NewConfig() (*Config, error) {
	cfg := new(Config)

	if err := envconfig.Init(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
