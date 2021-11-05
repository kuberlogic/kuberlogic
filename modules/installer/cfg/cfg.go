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
	"fmt"
	logger "github.com/kuberlogic/kuberlogic/modules/installer/log"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"os"
	"strings"
)

// default configuration variables
var (
	requiredParamNotSet = fmt.Errorf("some required parameter(s) not set")

	defaultKubeconfigPath   = fmt.Sprintf("%s/%s", os.Getenv("HOME"), ".kube/config")
	defaultDebugLogsEnabled = false
	defaultHelmReleaseName  = "kuberlogic"

	defaultPlatform    = "generic"
	supportedPlatforms = []string{defaultPlatform, "eks"}
)

type Config struct {
	DebugLogs      *bool   `yaml:"debugLogs,omitempty"`
	KubeconfigPath *string `yaml:"kubeconfigPath,omitempty"`

	Namespace *string `yaml:"namespace"`

	Endpoints struct {
		API               string `yaml:"api""`
		UI                string `yaml:"ui"`
		MonitoringConsole string `yaml:"monitoringConsole"`
	} `yaml:"endpoints"`

	Registry struct {
		Server   string `yaml:"server,omitempty"`
		Username string `yaml:"username,omitempty"`
		Password string `yaml:"password,omitempty"`
	} `yaml:"registry,omitempty"`

	Auth struct {
		AdminPassword    string  `yaml:"adminPassword"`
		DemoUserPassword *string `yaml:"demoUserPassword,omitempty"`
	} `yaml:"auth"`

	Platform string `yaml:"platform,omitempty"`
}

func (c *Config) setDefaults(log logger.Logger) error {
	var configError error
	if c.DebugLogs == nil {
		log.Debugf("Using default value for debugLogs: %s", defaultDebugLogsEnabled)
		v := &defaultDebugLogsEnabled
		c.DebugLogs = v
	}

	if c.KubeconfigPath == nil {
		log.Debugf("Using default value for kubeconfigPath: %s", defaultKubeconfigPath)
		v := &defaultKubeconfigPath
		c.KubeconfigPath = v
	}

	if c.Namespace == nil {
		log.Errorf("`namespace` config value can't be empty")
		configError = requiredParamNotSet
	}

	if c.Endpoints.UI == "" || c.Endpoints.API == "" {
		log.Errorf("`endpoints.api` and `endpoints.ui` must be set and can't be-empty")
		return errors.New("endpoints configuration is not set")
	}

	if c.Endpoints.MonitoringConsole == "" {
		log.Errorf("`endpoints.monitoringConsole` must be set and can't be empty")
		return errors.New("endpoints.monitoringConsole is not set")
	}

	if c.Platform == "" {
		log.Debugf("Using default value for platform: %s", defaultPlatform)
		c.Platform = defaultPlatform
	} else {
		matched := false
		for _, p := range supportedPlatforms {
			if strings.ToUpper(p) == strings.ToUpper(c.Platform) {
				matched = true
			}
		}
		if !matched {
			log.Errorf("Unsupported platform. List of supported platforms: %v", supportedPlatforms)
			return errors.New("unsupported platform")
		}
	}

	return configError
}

func NewConfigFromFile(file string, log logger.Logger) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg := new(Config)
	c := yaml.NewDecoder(f)
	if err := c.Decode(cfg); err != nil {
		return nil, err
	}

	if err := cfg.setDefaults(log); err != nil {
		return nil, err
	}
	return cfg, nil
}
