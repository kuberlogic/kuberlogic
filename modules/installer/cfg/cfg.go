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
	defaultPlatform         = "generic"
	supportedPlatforms      = []string{defaultPlatform, "aws"}
)

type TLS struct {
	CaFile  string `yaml:"ca.crt"`
	CrtFile string `yaml:"tls.crt"`
	KeyFile string `yaml:"tls.key"`
}

type Config struct {
	DebugLogs      *bool   `yaml:"debug-logs,omitempty"`
	KubeconfigPath *string `yaml:"kubeconfig-path,omitempty"`

	Namespace *string `yaml:"namespace"`

	Endpoints struct {
		Kuberlogic    string `yaml:"kuberlogic"`
		KuberlogicTLS *TLS   `yaml:"kuberlogic-tls,omitempty"`

		MonitoringConsole    string `yaml:"monitoring-console"`
		MonitoringConsoleTLS *TLS   `yaml:"monitoring-console-tls,omitempty"`
	} `yaml:"endpoints"`

	Registry struct {
		Server   string `yaml:"server,omitempty"`
		Username string `yaml:"username,omitempty"`
		Password string `yaml:"password,omitempty"`
	} `yaml:"registry,omitempty"`

	Auth struct {
		AdminPassword    string  `yaml:"admin-password"`
		DemoUserPassword *string `yaml:"demo-user-password,omitempty"`
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
		log.Debugf("Using default value for kubeconfig-path: %s", defaultKubeconfigPath)
		v := &defaultKubeconfigPath
		c.KubeconfigPath = v
	}

	if c.Namespace == nil {
		log.Errorf("`namespace` config value can't be empty")
		configError = requiredParamNotSet
	}

	if c.Endpoints.Kuberlogic == "" {
		log.Errorf("`endpoints.main` must be set and can't be-empty")
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

func (c *Config) checkKuberlogicTLS() error {
	if err := checkCertificates(c.Endpoints.KuberlogicTLS); err != nil {
		return err
	}

	if err := checkCertificates(c.Endpoints.MonitoringConsoleTLS); err != nil {
		return err
	}
	return nil
}

func (c *Config) check() error {
	return c.checkKuberlogicTLS()
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
	if err := cfg.check(); err != nil {
		return nil, err
	}
	return cfg, nil
}
