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
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

// default configuration variables
var (
	errRequiredParamNotSet = fmt.Errorf("some required parameter(s) not set")

	DefaultKubeconfigPath   = fmt.Sprintf("%s/%s", os.Getenv("HOME"), ".kube/config")
	defaultDebugLogsEnabled = false
	defaultSMTPFromUser     = "notifications"
	defaultPlatform         = "generic"
	supportedPlatforms      = []string{defaultPlatform, "aws"}
)

type TLS struct {
	CaFile  string `yaml:"ca.crt"`
	CrtFile string `yaml:"tls.crt"`
	KeyFile string `yaml:"tls.key"`
}

type SMTP struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	TLS      struct {
		Enabled  bool `yaml:"enabled"`
		Insecure bool `yaml:"insecure"`
	} `yaml:"tls,omitempty"`
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

	SMTP *SMTP `yaml:"smtp,omitempty"`

	Platform string `yaml:"platform,omitempty"`
}

func (c *Config) setDefaults(log logger.Logger) {
	if c.DebugLogs == nil {
		log.Debugf("Using default value for debugLogs: %s", defaultDebugLogsEnabled)
		c.DebugLogs = &defaultDebugLogsEnabled
	}

	if c.KubeconfigPath == nil {
		log.Debugf("Using default value for kubeconfig-path: %s", DefaultKubeconfigPath)
		c.KubeconfigPath = &DefaultKubeconfigPath
	}

	if c.Platform == "" {
		log.Debugf("Using default value for platform: %s", defaultPlatform)
		c.Platform = defaultPlatform
	}

	if c.SMTP != nil {
		defaultSMTPFrom := defaultSMTPFromUser + "@" + c.Endpoints.Kuberlogic
		log.Debugf("Using default value for `smtp.from`: %s", defaultSMTPFrom)
		c.SMTP.From = defaultSMTPFrom
	}
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

func (c *Config) checkPlatform() error {
	if c.Platform == "" {
		return nil
	}
	matched := false
	for _, p := range supportedPlatforms {
		if strings.ToUpper(p) == strings.ToUpper(c.Platform) {
			matched = true
		}
	}
	if !matched {
		return errors.New(fmt.Sprintf("platform is not in the supported list: %v", supportedPlatforms))
	}
	return nil
}

func (c *Config) checkSMTP() error {
	if c.SMTP == nil {
		return nil
	}
	basicErr := errors.New("`smtp` setting is not valid")
	if c.SMTP.Host == "" {
		return errors.Wrap(basicErr, "`host` must be set")
	}
	if c.SMTP.Port == 0 {
		return errors.Wrap(basicErr, "`port` is not correct or not set")
	}
	return nil
}

func (c *Config) check() error {
	if c.Endpoints.Kuberlogic == "" || c.Endpoints.MonitoringConsole == "" {
		return errors.Wrap(errRequiredParamNotSet, "`kuberlogic` and `monitoring-console` endpoints names must be configured")
	}

	if c.Namespace == nil {
		return errors.Wrap(errRequiredParamNotSet, "`namespace` must be configured")
	}

	if err := c.checkPlatform(); err != nil {
		return errors.Wrap(err, "error checking platform configuration")
	}

	if err := c.checkKuberlogicTLS(); err != nil {
		return errors.Wrap(err, "error checking TLS configuration")
	}
	return nil
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

	if err := cfg.check(); err != nil {
		return nil, err
	}
	cfg.setDefaults(log)
	return cfg, nil
}

func newFileFromConfig(cfg *Config, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return errors.Wrap(err, "cannot create config file")
	}
	defer f.Close()

	err = yaml.NewEncoder(f).Encode(cfg)
	if err != nil {
		return errors.Wrap(err, "cannot encode config to file")
	}
	return nil
}
