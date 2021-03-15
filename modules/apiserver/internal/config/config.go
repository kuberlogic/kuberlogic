package config

import (
	"github.com/kuberlogic/operator/modules/apiserver/internal/logging"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

// Config struct
type Config struct {
	BindHost     string `envconfig:"default=0.0.0.0"`
	HTTPBindPort int    `envconfig:"default=8001"`

	Auth struct {
		Provider string

		Keycloak struct {
			ClientId     string
			ClientSecret string
			RealmName    string
			Url          string
		} `envconfig:"optional"`
	}
	KubeconfigPath string `envconfig:"default=~/.kube/config"`
	DebugLogs      bool   `envconfig:"default=false"`
}

// InitConfig func
func InitConfig(prefix string, log logging.Logger) (*Config, error) {
	config := &Config{}
	if err := envconfig.InitWithPrefix(config, prefix); err != nil {
		return nil, errors.Wrap(err, "init config failed")
	}

	log.Debugf("config is %v", config)

	return config, nil
}
