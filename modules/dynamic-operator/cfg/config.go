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

type Config struct {
	// The address the metric endpoint binds to.
	MetricsAddr string `envconfig:"default=:8080,optional"`
	// The address the probe endpoint binds to
	ProbeAddr string `envconfig:"default=:8081,optional"`
	// Enable leader election for controller manager.
	// Enabling this will ensure there is only one active controller manager.
	EnableLeaderElection bool `envconfig:"default=false,optional"`

	Plugins []struct {
		Name string
		Path string
	}
}

func NewConfig() (*Config, error) {
	cfg := new(Config)
	if err := envconfig.Init(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
