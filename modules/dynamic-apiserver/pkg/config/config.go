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

package config

import (
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/logging"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

// Config struct
type Config struct {
	BindHost     string `envconfig:"default=0.0.0.0"`
	HTTPBindPort int    `envconfig:"default=8001"`

	KubeconfigPath string `envconfig:"default=/root/.kube/config"`
	DebugLogs      bool   `envconfig:"default=false"`
	Cors           struct {
		AllowedOrigins []string `envconfig:"default=https://*;http://*"`
	}
	SentryDsn string `envconfig:"optional"`
	Domain    string
}

// InitConfig func
func InitConfig(prefix string, log logging.Logger) (*Config, error) {
	config := &Config{}
	if err := envconfig.InitWithPrefix(config, prefix); err != nil {
		return nil, errors.Wrap(err, "init config failed")
	}

	log.Debugw("config is", "config", config)
	return config, nil
}
