/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cfg

import (
	"github.com/vrischmann/envconfig"
)

type Grafana struct {
	Enabled  bool   `envconfig:"default=false,optional"`
	Endpoint string `envconfig:"optional"`
	Login    string `envconfig:"default=admin,optional"`
	Password string `envconfig:"default=admin,optional"`

	DefaultDatasourceEndpoint string `envconfig:"optional"`
}

type Config struct {
	MetricsAddr          string `envconfig:"default=:8080,optional"`
	EnableLeaderElection bool   `envconfig:"default=false,optional"`

	ImageRepo           string `envconfig:"IMG_REPO"`
	ImagePullSecretName string `envconfig:"IMG_PULL_SECRET,optional"`
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
