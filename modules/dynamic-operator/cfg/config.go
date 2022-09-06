/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package cfg

import (
	"github.com/vrischmann/envconfig"
)

type Config struct {
	// Namespace where controller is running
	Namespace string `envconfig:""`
	// ServiceAccount of controller
	ServiceAccount string `envconfig:"optional"`
	// Provisioned services will use this IngressClass / StorageClass or default (if empty)
	IngressClass string `envconfig:"optional"`
	StorageClass string `envconfig:"optional"`

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

	DeploymentId string `envconfig:""`
	SentryDsn    string `envconfig:"optional"`
}

func NewConfig() (*Config, error) {
	cfg := new(Config)
	if err := envconfig.Init(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
