/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package cfg

import (
	"github.com/vrischmann/envconfig"
)

type Config struct {
	// The address the metric endpoint binds to.
	MetricsAddr string `envconfig:"default=:8080,optional"`
	// The address the probe endpoint binds to
	ProbeAddr string `envconfig:"default=:8081,optional"`
	// Enable leader election for controller manager.
	// Enabling this will ensure there is only one active controller manager.
	EnableLeaderElection bool `envconfig:"default=false,optional"`
	// Namespace where controller is running
	Namespace string `envconfig:""`
	// ServiceAccount of controller
	ServiceAccount string `envconfig:"optional"`

	Plugins []struct {
		Name string
		Path string
	} `envconfig:"optional"`

	// additional options for service environment configuration
	SvcOpts struct {
		TLSSecretName string `envconfig:"optional"`
	} `envconfig:"optional"`

	Backups struct {
		Enabled          bool `enconfig:"default=false,optional"`
		SnapshotsEnabled bool `envconfig:"optional"`
	} `envconfig:"optional"`

	SentryDsn string `envconfig:"optional"`
}

func NewConfig() (*Config, error) {
	cfg := new(Config)
	if err := envconfig.Init(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
